package clct

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/prutonis/acquisitor/internal/cfg"
	"github.com/prutonis/acquisitor/internal/modbus"
	"github.com/prutonis/acquisitor/pkg/logger"
)

// Solar charge controller read over Modbus RTU (see ChargeController-modbus.pdf).
//
// Everything is configured in the yaml file:
//   hardware.charger.registers : which registers exist (address, type, shift, mask)
//   telemetry.collectors[charger].keys : what is published, with
//       source   = name of a register
//       function = CEL expression on `raw` (optional), e.g. "raw / 10.0"
//       values / bits / empty = optional text mapping of the result

const (
	COLLECTOR_NAME_CHARGER = "charger"

	// source "ok": 1 when the controller answered in the last cycle, 0 otherwise
	chargerSourceOK = "ok"
	// after this many consecutive failed cycles the old values are dropped,
	// so stale data is not published as if it were fresh
	chargerMaxFailures = 3
	chargerInfoTries   = 3
	chargerMaxBlock    = 120 // registers per request (protocol limit is 125)
)

// regReader is what the collector needs from the modbus client (and from the simulator).
type regReader interface {
	ReadHoldingRegisters(slave byte, addr, qty uint16) ([]uint16, error)
	Close() error
}

type regMap map[uint16]uint16

// ---------------------------------------------------------------------------
// registers

// Address segments of the manual (section 2.3): they must not be crossed by one request.
var chargerSegments = [...][2]uint16{
	{0x0000, 0x0009},
	{0x000A, 0x001A},
	{0x0100, 0x0122},
	{0xE001, 0xE02D},
	{0xF000, 0xF3FF},
}

func chargerSegment(a uint16) int {
	for i, s := range chargerSegments {
		if a >= s[0] && a <= s[1] {
			return i
		}
	}
	return 100 + int(a) // unknown area: never merged with other addresses
}

type chargerRegister struct {
	name   string
	addr   uint16
	words  int // 1 or 2
	signed bool
	shift  uint
	mask   uint32
}

func newChargerRegister(c cfg.ChargerRegister) (*chargerRegister, error) {
	if c.Name == "" {
		return nil, errors.New("register without name")
	}
	r := &chargerRegister{name: c.Name, words: 1, shift: c.Shift, mask: c.Mask}
	switch strings.ToLower(c.Type) {
	case "", "u16":
	case "i16":
		r.signed = true
	case "u32":
		r.words = 2
	case "i32":
		r.words = 2
		r.signed = true
	default:
		return nil, fmt.Errorf("register '%s': unknown type '%s' (u16, i16, u32, i32)", c.Name, c.Type)
	}
	if c.Address < 0 || c.Address+r.words-1 > 0xFFFF {
		return nil, fmt.Errorf("register '%s': invalid address %d", c.Name, c.Address)
	}
	if r.shift > 31 {
		return nil, fmt.Errorf("register '%s': invalid shift %d", c.Name, r.shift)
	}
	r.addr = uint16(c.Address)
	return r, nil
}

func (r *chargerRegister) span() (uint16, uint16) {
	return r.addr, r.addr + uint16(r.words) - 1
}

// raw returns the value of the register; false if one of its words was not read.
func (r *chargerRegister) raw(regs regMap) (float64, bool) {
	var v uint32
	for i := 0; i < r.words; i++ {
		w, ok := regs[r.addr+uint16(i)]
		if !ok {
			return 0, false
		}
		v = v<<16 | uint32(w)
	}
	if r.shift == 0 && r.mask == 0 && r.signed {
		if r.words == 1 {
			return float64(int16(v)), true
		}
		return float64(int32(v)), true
	}
	v >>= r.shift
	if r.mask != 0 {
		v &= r.mask
	}
	return float64(v), true
}

// ---------------------------------------------------------------------------
// read plan

type chargerBlock struct {
	start, count uint16
}

func sameChargerSegment(addrs ...uint16) bool {
	s := chargerSegment(addrs[0])
	for _, a := range addrs[1:] {
		if chargerSegment(a) != s {
			return false
		}
	}
	return true
}

// planChargerBlocks groups the registers in as few requests as possible:
// registers at most maxGap words apart, in the same address segment, are merged.
func planChargerBlocks(regs []*chargerRegister, maxGap int) []chargerBlock {
	type span struct{ s, e uint16 }
	spans := make([]span, 0, len(regs))
	for _, r := range regs {
		s, e := r.span()
		spans = append(spans, span{s, e})
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].s < spans[j].s })

	var blocks []chargerBlock
	for _, sp := range spans {
		if n := len(blocks); n > 0 {
			last := &blocks[n-1]
			lastEnd := last.start + last.count - 1
			newEnd := lastEnd
			if sp.e > newEnd {
				newEnd = sp.e
			}
			if int(sp.s) <= int(lastEnd)+1+maxGap &&
				sameChargerSegment(last.start, lastEnd, sp.s, sp.e) &&
				int(newEnd)-int(last.start)+1 <= chargerMaxBlock {
				last.count = newEnd - last.start + 1
				continue
			}
		}
		blocks = append(blocks, chargerBlock{start: sp.s, count: sp.e - sp.s + 1})
	}
	return blocks
}

// ---------------------------------------------------------------------------
// keys

type chargerBit struct {
	bit  uint
	name string
}

type chargerKey struct {
	name   string
	reg    *chargerRegister // nil: the "ok" key
	prog   cel.Program
	values map[int]string
	bits   []chargerBit
	empty  string
}

func newChargerKey(k cfg.CollectorKey, reg *chargerRegister) (*chargerKey, error) {
	ck := &chargerKey{name: k.Name, reg: reg, prog: loadTransformFunction(k.Function), empty: k.Empty}
	if ck.empty == "" {
		ck.empty = "ok"
	}
	if len(k.Values) > 0 {
		ck.values = make(map[int]string, len(k.Values))
		for s, name := range k.Values {
			i, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil {
				return nil, fmt.Errorf("key '%s': values: '%s' is not a number", k.Name, s)
			}
			ck.values[i] = name
		}
	}
	if len(k.Bits) > 0 {
		for s, name := range k.Bits {
			b, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil || b < 0 || b > 63 {
				return nil, fmt.Errorf("key '%s': bits: '%s' is not a valid bit number", k.Name, s)
			}
			ck.bits = append(ck.bits, chargerBit{bit: uint(b), name: name})
		}
		sort.Slice(ck.bits, func(i, j int) bool { return ck.bits[i].bit < ck.bits[j].bit })
	}
	return ck, nil
}

// eval computes the telemetry value (float64 or string) of the key.
func (ck *chargerKey) eval(regs regMap) (interface{}, bool) {
	raw, ok := ck.reg.raw(regs)
	if !ok {
		return nil, false
	}
	res, _, err := ck.prog.Eval(map[string]interface{}{"raw": raw})
	if err != nil {
		logger.Debugf("Charge controller: key '%s': function failed: %v", ck.name, err)
		return nil, false
	}
	var f float64
	switch v := res.Value().(type) {
	case float64:
		f = v
	case int64:
		f = float64(v)
	case uint64:
		f = float64(v)
	case bool:
		if v {
			f = 1
		}
	case string:
		return v, true
	default:
		logger.Debugf("Charge controller: key '%s': unsupported result type %T", ck.name, v)
		return nil, false
	}

	switch {
	case ck.bits != nil:
		n := int64(math.Round(f))
		if n < 0 {
			n = 0
		}
		var names []string
		for _, b := range ck.bits {
			if uint64(n)&(1<<b.bit) != 0 {
				names = append(names, b.name)
			}
		}
		if len(names) == 0 {
			return ck.empty, true
		}
		return strings.Join(names, ","), true
	case ck.values != nil:
		n := int(math.Round(f))
		if name, ok := ck.values[n]; ok {
			return name, true
		}
		return fmt.Sprintf("unknown_%d", n), true
	}
	return f, true
}

// ---------------------------------------------------------------------------
// collector

type chargerCollector struct {
	enabled bool
	device  string
	baud    int
	slave   byte
	timeout time.Duration

	keys   []*chargerKey
	blocks []chargerBlock

	reader    regReader
	fails     int
	infoDone  bool
	infoTries int
}

func (cc *chargerCollector) Name() string {
	return COLLECTOR_NAME_CHARGER
}

func (cc *chargerCollector) Init() {
	cc.enabled = false
	hw := config.Hardware.Charger
	if !hw.Enabled {
		logger.Infof("Charge controller disabled")
		return
	}
	col := config.Telemetry.ResolveCollector(cc.Name())
	if col == nil {
		logger.Infof("Charge controller collector not configured")
		return
	}

	cc.device = hw.Device
	if cc.device == "" {
		cc.device = "/dev/ttyS2"
	}
	cc.baud = hw.Baud
	if cc.baud == 0 {
		cc.baud = 9600
	}
	cc.slave = 1
	if hw.SlaveId > 0 && hw.SlaveId <= 255 {
		cc.slave = byte(hw.SlaveId)
	}
	cc.timeout = time.Second
	if hw.Timeout > 0 {
		cc.timeout = time.Duration(hw.Timeout) * time.Millisecond
	}

	regDefs := make(map[string]*chargerRegister)
	for _, rc := range hw.Registers {
		r, err := newChargerRegister(rc)
		if err != nil {
			logger.Errorf("Charge controller: %v (ignored)", err)
			continue
		}
		regDefs[r.name] = r
	}

	cc.keys = nil
	var used []*chargerRegister
	for _, key := range col.Keys {
		if key.Source == chargerSourceOK {
			cc.keys = append(cc.keys, &chargerKey{name: key.Name})
			telemetryData.Init(key.Name, key.Unit, TYPE_FLOAT, false)
			continue
		}
		reg, ok := regDefs[key.Source]
		if !ok {
			logger.Errorf("Charge controller: key '%s': unknown register '%s' (ignored)", key.Name, key.Source)
			continue
		}
		ck, err := newChargerKey(key, reg)
		if err != nil {
			logger.Errorf("Charge controller: %v (ignored)", err)
			continue
		}
		cc.keys = append(cc.keys, ck)
		used = append(used, reg)
		telemetryData.Init(key.Name, key.Unit, TYPE_FLOAT, false)
	}

	cc.blocks = planChargerBlocks(used, hw.MaxGap)
	if len(cc.keys) == 0 {
		logger.Infof("Charge controller: no valid keys configured")
		return
	}
	cc.enabled = true
	logger.Infof("Charge controller: %s %d 8N1, slave %d, %d keys, %d read requests per cycle",
		cc.device, cc.baud, cc.slave, len(cc.keys), len(cc.blocks))
	for _, b := range cc.blocks {
		logger.Infof("  read %04XH, %d registers", b.start, b.count)
	}
}

func (cc *chargerCollector) Close() {
	if cc.reader != nil {
		cc.reader.Close()
		cc.reader = nil
	}
}

func (cc *chargerCollector) ensureReader() error {
	if cc.reader != nil {
		return nil
	}
	if strings.Contains(cc.device, "fake") {
		logger.Infof("Using fake charge controller")
		cc.reader = fakeChargerReader{}
		return nil
	}
	conn, err := modbus.OpenSerial(cc.device, cc.baud)
	if err != nil {
		return err
	}
	cc.reader = modbus.NewClient(conn, cc.timeout)
	logger.Infof("Charge controller: port %s opened", cc.device)
	return nil
}

func (cc *chargerCollector) Collect() {
	if !cc.enabled {
		return
	}
	logger.Debugf("Collecting charge controller data")

	if err := cc.ensureReader(); err != nil {
		cc.cycleFailed(err)
		return
	}

	regs := regMap{}
	anyOK := false
	var lastErr error

	for _, b := range cc.blocks {
		vals, err := cc.reader.ReadHoldingRegisters(cc.slave, b.start, b.count)
		if err != nil {
			lastErr = err
			logger.Debugf("Charge controller: read %04X/%d failed: %v", b.start, b.count, err)
			var exc *modbus.ExceptionError
			if errors.As(err, &exc) {
				// the controller answered but refuses this block, try the others
				continue
			}
			if errors.Is(err, modbus.ErrIO) {
				// serial port problem: reopen it at the next cycle
				cc.Close()
			}
			break // timeout / CRC: do not wait again for every block
		}
		for j, v := range vals {
			regs[b.start+uint16(j)] = v
		}
		anyOK = true
	}

	if !anyOK {
		cc.cycleFailed(lastErr)
		return
	}
	if cc.fails > 0 {
		logger.Infof("Charge controller: communication restored after %d failed cycles", cc.fails)
	}
	cc.fails = 0

	for _, key := range cc.keys {
		if key.reg == nil {
			telemetryData.AddValue(key.name, 1.0)
			continue
		}
		if v, ok := key.eval(regs); ok {
			telemetryData.AddValue(key.name, v)
		}
	}

	if !cc.infoDone && cc.infoTries < chargerInfoTries && cc.reader != nil {
		cc.logDeviceInfo()
	}
}

func (cc *chargerCollector) cycleFailed(err error) {
	cc.fails++
	if cc.fails == 1 || cc.fails%60 == 0 {
		logger.Errorf("Charge controller: read failed (%d consecutive): %v", cc.fails, err)
	}
	for _, key := range cc.keys {
		if key.reg == nil {
			telemetryData.AddValue(key.name, 0.0)
			continue
		}
		if cc.fails >= chargerMaxFailures {
			// value nil => not published by the pusher
			if tc := telemetryData.ResolveChannel(key.name); tc != nil {
				tc.Value = nil
			}
		}
	}
}

// logDeviceInfo reads the static data (000AH-0019H) once and writes it to the log.
func (cc *chargerCollector) logDeviceInfo() {
	cc.infoTries++
	v, err := cc.reader.ReadHoldingRegisters(cc.slave, 0x000A, 16)
	if err != nil {
		logger.Infof("Charge controller: cannot read device info: %v", err)
		return
	}
	cc.infoDone = true

	model := make([]byte, 0, 16)
	for _, w := range v[2:10] {
		model = append(model, byte(w>>8), byte(w))
	}
	logger.Infof("Charge controller: model '%s', system %dV, rated charge %dA, rated discharge %dA, type %d",
		strings.Trim(string(model), " \x00"), v[0]>>8, v[0]&0xFF, v[1]>>8, v[1]&0xFF)
	logger.Infof("Charge controller: software V%02d.%02d.%02d, hardware V%02d.%02d.%02d, serial %04X%04X",
		v[10]&0xFF, v[11]>>8, v[11]&0xFF, v[12]&0xFF, v[13]>>8, v[13]&0xFF, v[14], v[15])
}

// ---------------------------------------------------------------------------
// Simulator, used when hardware.charger.device contains "fake" (development).

type fakeChargerReader struct{}

func (fakeChargerReader) Close() error { return nil }

func (fakeChargerReader) ReadHoldingRegisters(slave byte, addr, qty uint16) ([]uint16, error) {
	out := make([]uint16, qty)
	for i := range out {
		out[i] = fakeChargerRegister(addr + uint16(i))
	}
	return out, nil
}

func fakeChargerRegister(a uint16) uint16 {
	switch a {
	case 0x000A:
		return 0x180A // 24V, 10A
	case 0x000B:
		return 0x0A00
	case 0x0100:
		return 87
	case 0x0101:
		return uint16(250 + rand.Intn(20)) // 25.0 - 26.9 V
	case 0x0102:
		return uint16(100 + rand.Intn(300)) // 1.00 - 3.99 A
	case 0x0103:
		return 0x1C19 // controller 28C, battery 25C
	case 0x0104:
		return 251
	case 0x0105:
		return 120
	case 0x0106:
		return 30
	case 0x0107:
		return uint16(300 + rand.Intn(50))
	case 0x0108:
		return 230
	case 0x0109:
		return 70
	case 0x0120:
		return 0xE402 // load on, 100%, mppt
	}
	return 0
}
