package modbus

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"
)

func mustHex(s string) []byte {
	b, err := hex.DecodeString(strings.ReplaceAll(s, " ", ""))
	if err != nil {
		panic(err)
	}
	return b
}

// fakeConn answers with the given bytes, one byte per Read, and every second
// Read returns (0, nil) like a real port does on timeout.
type fakeConn struct {
	resp    []byte
	pos     int
	tick    bool
	written bytes.Buffer
}

func (f *fakeConn) Read(p []byte) (int, error) {
	f.tick = !f.tick
	if f.tick || f.pos >= len(f.resp) {
		return 0, nil
	}
	p[0] = f.resp[f.pos]
	f.pos++
	return 1, nil
}
func (f *fakeConn) Write(p []byte) (int, error) { return f.written.Write(p) }
func (f *fakeConn) Flush() error                { return nil }
func (f *fakeConn) Close() error                { return nil }

func TestCRC16(t *testing.T) {
	// frames taken from the controller manual (CRC sent low byte first)
	cases := []struct{ data, crc string }{
		{"01 03 000A 0001", "A408"},
		{"01 03 02 181E", "324C"},
		{"01 03 0100 0001", "85F6"},
		{"01 03 02 0064", "B9AF"},
		{"01 03 0121 0002", "95FD"},
		{"01 03 04 0000 0021", "3A2B"},
	}
	for _, c := range cases {
		got := CRC16(mustHex(c.data))
		want := mustHex(c.crc)
		if byte(got) != want[0] || byte(got>>8) != want[1] {
			t.Errorf("CRC(%s) = %02X%02X, want %s", c.data, byte(got), byte(got>>8), c.crc)
		}
	}
}

func TestReadHoldingRegisters(t *testing.T) {
	conn := &fakeConn{resp: mustHex("01 03 02 181E 324C")}
	cl := NewClient(conn, time.Second)
	regs, err := cl.ReadHoldingRegisters(1, 0x000A, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(regs) != 1 || regs[0] != 0x181E {
		t.Fatalf("unexpected registers %v", regs)
	}
	if want := mustHex("01 03 000A 0001 A408"); !bytes.Equal(conn.written.Bytes(), want) {
		t.Errorf("request = %X, want %X", conn.written.Bytes(), want)
	}
}

func TestReadHoldingRegistersBadCRC(t *testing.T) {
	cl := NewClient(&fakeConn{resp: mustHex("01 03 02 181E 0000")}, time.Second)
	if _, err := cl.ReadHoldingRegisters(1, 0x000A, 1); !errors.Is(err, ErrCRC) {
		t.Fatalf("expected ErrCRC, got %v", err)
	}
}

func TestReadHoldingRegistersException(t *testing.T) {
	frame := mustHex("01 83 02")
	crc := CRC16(frame)
	frame = append(frame, byte(crc), byte(crc>>8))
	cl := NewClient(&fakeConn{resp: frame}, time.Second)
	_, err := cl.ReadHoldingRegisters(1, 0x0500, 1)
	var exc *ExceptionError
	if !errors.As(err, &exc) || exc.Code != 2 {
		t.Fatalf("expected exception 2, got %v", err)
	}
}

func TestReadHoldingRegistersTimeout(t *testing.T) {
	cl := NewClient(&fakeConn{}, 100*time.Millisecond)
	if _, err := cl.ReadHoldingRegisters(1, 0x000A, 1); !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got %v", err)
	}
}