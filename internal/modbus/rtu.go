// Package modbus contains a minimal Modbus RTU master (function 0x03 only),
// enough to read the registers of the solar charge controller.
package modbus

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	funcReadHolding = 0x03
	maxRegisters    = 125
	// minimal silence between two frames (the controller manual asks for >10ms)
	interFrameGap = 20 * time.Millisecond
)

var (
	// ErrTimeout is returned when the device does not answer in time.
	ErrTimeout = errors.New("modbus: response timeout")
	// ErrCRC is returned when the CRC of the response is not correct.
	ErrCRC = errors.New("modbus: bad CRC")
	// ErrIO is returned (wrapped) for errors of the underlying serial port.
	ErrIO = errors.New("modbus: i/o error")
	// ErrFrame is returned when the response is malformed.
	ErrFrame = errors.New("modbus: malformed response")
)

// ExceptionError is a Modbus exception response (function code | 0x80).
type ExceptionError struct {
	Function byte
	Code     byte
}

func (e *ExceptionError) Error() string {
	return fmt.Sprintf("modbus: exception 0x%02X for function 0x%02X", e.Code, e.Function)
}

// Conn is the serial line abstraction.
// Read must NOT block forever: when nothing arrives within a short time
// (e.g. 100ms) it returns (0, nil).
type Conn interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	// Flush discards unread input bytes.
	Flush() error
	Close() error
}

// Client is a Modbus RTU master. It is safe for concurrent use.
type Client struct {
	mu      sync.Mutex
	conn    Conn
	timeout time.Duration
	gap     time.Duration
	lastIO  time.Time
}

// NewClient creates a client on top of conn. timeout is the maximum time
// to wait for a complete response.
func NewClient(conn Conn, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = time.Second
	}
	return &Client{conn: conn, timeout: timeout, gap: interFrameGap}
}

// Close closes the underlying port.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.Close()
}

// CRC16 computes the Modbus CRC (poly 0xA001, init 0xFFFF).
func CRC16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

// ReadHoldingRegisters reads qty registers starting at addr (function 0x03).
// With slave 0xFF (the "bypass address" of the controller) the address byte
// of the response is not checked.
func (c *Client) ReadHoldingRegisters(slave byte, addr, qty uint16) ([]uint16, error) {
	if qty < 1 || qty > maxRegisters {
		return nil, fmt.Errorf("modbus: invalid register count %d", qty)
	}

	req := make([]byte, 8)
	req[0] = slave
	req[1] = funcReadHolding
	binary.BigEndian.PutUint16(req[2:], addr)
	binary.BigEndian.PutUint16(req[4:], qty)
	crc := CRC16(req[:6])
	req[6] = byte(crc) // CRC: low byte first
	req[7] = byte(crc >> 8)

	c.mu.Lock()
	defer c.mu.Unlock()

	if wait := c.gap - time.Since(c.lastIO); wait > 0 {
		time.Sleep(wait)
	}
	if err := c.conn.Flush(); err != nil {
		return nil, fmt.Errorf("%w: flush: %v", ErrIO, err)
	}
	if err := writeAll(c.conn, req); err != nil {
		return nil, fmt.Errorf("%w: write: %v", ErrIO, err)
	}

	resp, err := c.readResponse(qty)
	c.lastIO = time.Now()
	if err != nil {
		return nil, err
	}

	if slave != 0xFF && resp[0] != slave {
		return nil, fmt.Errorf("%w: unexpected slave address 0x%02X", ErrFrame, resp[0])
	}
	if resp[1] == funcReadHolding|0x80 {
		return nil, &ExceptionError{Function: funcReadHolding, Code: resp[2]}
	}
	if resp[1] != funcReadHolding {
		return nil, fmt.Errorf("%w: unexpected function 0x%02X", ErrFrame, resp[1])
	}

	regs := make([]uint16, qty)
	for i := range regs {
		regs[i] = binary.BigEndian.Uint16(resp[3+2*i:])
	}
	return regs, nil
}

// readResponse collects bytes until a complete frame (normal or exception)
// is received, then validates its CRC. It returns the whole frame.
func (c *Client) readResponse(qty uint16) ([]byte, error) {
	deadline := time.Now().Add(c.timeout)
	buf := make([]byte, 0, 5+2*maxRegisters)
	tmp := make([]byte, 256)
	need := 5 // shortest valid frame: the exception response

	for {
		if len(buf) >= 2 && buf[1]&0x80 != 0 {
			need = 5
		} else if len(buf) >= 3 {
			if int(buf[2]) != 2*int(qty) {
				return nil, fmt.Errorf("%w: byte count %d, expected %d", ErrFrame, buf[2], 2*int(qty))
			}
			need = 5 + int(buf[2])
		}
		if len(buf) >= need {
			break
		}
		if time.Now().After(deadline) {
			if len(buf) == 0 {
				return nil, ErrTimeout
			}
			return nil, fmt.Errorf("%w: incomplete frame (%d bytes)", ErrFrame, len(buf))
		}
		n, err := c.conn.Read(tmp)
		if err != nil {
			return nil, fmt.Errorf("%w: read: %v", ErrIO, err)
		}
		buf = append(buf, tmp[:n]...)
	}

	frame := buf[:need]
	got := uint16(frame[need-2]) | uint16(frame[need-1])<<8
	if CRC16(frame[:need-2]) != got {
		return nil, ErrCRC
	}
	return frame, nil
}

func writeAll(conn Conn, data []byte) error {
	for len(data) > 0 {
		n, err := conn.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}
