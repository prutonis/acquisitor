package clct

import (
	"reflect"
	"testing"

	"github.com/prutonis/acquisitor/internal/cfg"
)

func testChargerValue(t *testing.T, reg cfg.ChargerRegister, key cfg.CollectorKey, regs regMap) interface{} {
	t.Helper()
	r, err := newChargerRegister(reg)
	if err != nil {
		t.Fatal(err)
	}
	k, err := newChargerKey(key, r)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := k.eval(regs)
	if !ok {
		t.Fatalf("key %s: no value", key.Name)
	}
	return v
}

// values from the examples of the controller manual (sections 3 and 4)
func TestChargerScaling(t *testing.T) {
	regs := regMap{0x0100: 0x0037, 0x0101: 0x007B, 0x0102: 0x010A, 0x0105: 0x040B}
	cases := []struct {
		reg  cfg.ChargerRegister
		key  cfg.CollectorKey
		want float64
	}{
		{cfg.ChargerRegister{Name: "soc", Address: 0x0100, Mask: 0xFF}, cfg.CollectorKey{Name: "soc"}, 55},
		{cfg.ChargerRegister{Name: "bv", Address: 0x0101}, cfg.CollectorKey{Name: "bat_v", Function: "raw * 0.1"}, 12.3},
		{cfg.ChargerRegister{Name: "ci", Address: 0x0102}, cfg.CollectorKey{Name: "chg_i", Function: "raw / 100.0"}, 2.66},
		{cfg.ChargerRegister{Name: "li", Address: 0x0105}, cfg.CollectorKey{Name: "load_i", Function: "raw / 100.0"}, 10.35},
	}
	for _, c := range cases {
		got := testChargerValue(t, c.reg, c.key, regs).(float64)
		if diff := got - c.want; diff > 1e-9 || diff < -1e-9 {
			t.Errorf("%s = %v, want %v", c.key.Name, got, c.want)
		}
	}
}

func TestChargerTemperatureBytes(t *testing.T) {
	fn := "raw >= 128.0 ? -(raw - 128.0) : raw"
	ctrl := cfg.ChargerRegister{Name: "tc", Address: 0x0103, Shift: 8, Mask: 0xFF}
	bat := cfg.ChargerRegister{Name: "tb", Address: 0x0103, Mask: 0xFF}
	key := cfg.CollectorKey{Name: "t", Function: fn}

	regs := regMap{0x0103: 0x1C19} // manual 4.7: controller 28C, battery 25C
	if got := testChargerValue(t, ctrl, key, regs).(float64); got != 28 {
		t.Errorf("controller = %v", got)
	}
	if got := testChargerValue(t, bat, key, regs).(float64); got != 25 {
		t.Errorf("battery = %v", got)
	}
	regs = regMap{0x0103: 0x8385} // -3C / -5C
	if got := testChargerValue(t, ctrl, key, regs).(float64); got != -3 {
		t.Errorf("controller = %v", got)
	}
	if got := testChargerValue(t, bat, key, regs).(float64); got != -5 {
		t.Errorf("battery = %v", got)
	}
}

func TestChargerStatusText(t *testing.T) {
	regs := regMap{0x0120: 0xE402, 0x0121: 0x0000, 0x0122: 0x0021}

	on := cfg.ChargerRegister{Name: "on", Address: 0x0120, Shift: 15, Mask: 1}
	if got := testChargerValue(t, on, cfg.CollectorKey{Name: "load_on"}, regs).(float64); got != 1 {
		t.Errorf("load_on = %v", got)
	}
	br := cfg.ChargerRegister{Name: "br", Address: 0x0120, Shift: 8, Mask: 0x7F}
	if got := testChargerValue(t, br, cfg.CollectorKey{Name: "b"}, regs).(float64); got != 100 {
		t.Errorf("brightness = %v", got)
	}

	state := cfg.ChargerRegister{Name: "st", Address: 0x0120, Mask: 0xFF}
	stKey := cfg.CollectorKey{Name: "s", Values: map[string]string{"1": "activated", "2": "mppt"}}
	if got := testChargerValue(t, state, stKey, regs).(string); got != "mppt" {
		t.Errorf("state = %v", got)
	}

	// manual 4.15: 0021H -> b5 controller too hot, b0 battery over-discharge
	fault := cfg.ChargerRegister{Name: "f", Address: 0x0121, Type: "u32"}
	fKey := cfg.CollectorKey{Name: "ft", Bits: map[string]string{
		"0": "battery_overdischarge", "5": "controller_overtemp", "22": "ignored_unless_set",
	}}
	if got := testChargerValue(t, fault, fKey, regs).(string); got != "battery_overdischarge,controller_overtemp" {
		t.Errorf("fault_text = %v", got)
	}
	regs[0x0122] = 0
	if got := testChargerValue(t, fault, fKey, regs).(string); got != "ok" {
		t.Errorf("fault_text = %v, want ok", got)
	}
}

func TestChargerWideAndSigned(t *testing.T) {
	// manual 4.12: 00010203H = 66051 Ah
	regs := regMap{0x0118: 0x0001, 0x0119: 0x0203}
	tot := cfg.ChargerRegister{Name: "tot", Address: 0x0118, Type: "u32"}
	if got := testChargerValue(t, tot, cfg.CollectorKey{Name: "t"}, regs).(float64); got != 66051 {
		t.Errorf("total = %v", got)
	}
	regs = regMap{0x0200: 0xFFFE}
	s := cfg.ChargerRegister{Name: "s", Address: 0x0200, Type: "i16"}
	if got := testChargerValue(t, s, cfg.CollectorKey{Name: "s"}, regs).(float64); got != -2 {
		t.Errorf("signed = %v", got)
	}
}

func TestChargerMissingRegister(t *testing.T) {
	r, _ := newChargerRegister(cfg.ChargerRegister{Name: "x", Address: 0x0101})
	k, _ := newChargerKey(cfg.CollectorKey{Name: "x"}, r)
	if _, ok := k.eval(regMap{}); ok {
		t.Error("expected no value when the register was not read")
	}
}

func TestPlanChargerBlocks(t *testing.T) {
	mk := func(addrs ...int) []*chargerRegister {
		var out []*chargerRegister
		for _, a := range addrs {
			r, err := newChargerRegister(cfg.ChargerRegister{Name: "r", Address: a})
			if err != nil {
				t.Fatal(err)
			}
			out = append(out, r)
		}
		return out
	}
	cases := []struct {
		name string
		regs []*chargerRegister
		gap  int
		want []chargerBlock
	}{
		{"adjacent", mk(0x101, 0x100, 0x102), 0, []chargerBlock{{0x100, 3}}},
		{"gap not allowed", mk(0x100, 0x102), 0, []chargerBlock{{0x100, 1}, {0x102, 1}}},
		{"gap allowed", mk(0x100, 0x102), 1, []chargerBlock{{0x100, 3}}},
		{"segments not crossed", mk(0x009, 0x00A), 5, []chargerBlock{{0x009, 1}, {0x00A, 1}}},
		{"same address twice", mk(0x103, 0x103), 0, []chargerBlock{{0x103, 1}}},
	}
	for _, c := range cases {
		if got := planChargerBlocks(c.regs, c.gap); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
	// a 32 bit register takes two words
	r, _ := newChargerRegister(cfg.ChargerRegister{Name: "w", Address: 0x118, Type: "u32"})
	if got := planChargerBlocks([]*chargerRegister{r}, 0); !reflect.DeepEqual(got, []chargerBlock{{0x118, 2}}) {
		t.Errorf("u32: got %v", got)
	}
}
