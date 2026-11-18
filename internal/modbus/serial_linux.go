//go:build linux

package modbus

import (
	"fmt"

	"golang.org/x/sys/unix"
)

type serialPort struct {
	fd int
}

var baudRates = map[int]uint32{
	1200:   unix.B1200,
	2400:   unix.B2400,
	4800:   unix.B4800,
	9600:   unix.B9600,
	19200:  unix.B19200,
	38400:  unix.B38400,
	57600:  unix.B57600,
	115200: unix.B115200,
}

// OpenSerial opens a serial port in raw mode, 8 data bits, no parity, 1 stop bit.
// Read() returns (0, nil) after ~100ms without data.
func OpenSerial(device string, baud int) (Conn, error) {
	speed, ok := baudRates[baud]
	if !ok {
		return nil, fmt.Errorf("unsupported baud rate %d", baud)
	}
	fd, err := unix.Open(device, unix.O_RDWR|unix.O_NOCTTY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", device, err)
	}

	t, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("tcgetattr %s: %w", device, err)
	}

	t.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR |
		unix.IGNCR | unix.ICRNL | unix.IXON | unix.IXOFF | unix.IXANY | unix.INPCK
	t.Oflag &^= unix.OPOST
	t.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	t.Cflag &^= unix.CSIZE | unix.PARENB | unix.CSTOPB | unix.CRTSCTS | unix.CBAUD
	t.Cflag |= unix.CS8 | unix.CREAD | unix.CLOCAL | speed
	t.Cc[unix.VMIN] = 0  // return even with no data...
	t.Cc[unix.VTIME] = 1 // ...after 100 ms

	if err := unix.IoctlSetTermios(fd, unix.TCSETS, t); err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("tcsetattr %s: %w", device, err)
	}
	return &serialPort{fd: fd}, nil
}

func (p *serialPort) Read(b []byte) (int, error) {
	for {
		n, err := unix.Read(p.fd, b)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return 0, err
		}
		return n, nil
	}
}

func (p *serialPort) Write(b []byte) (int, error) {
	total := 0
	for total < len(b) {
		n, err := unix.Write(p.fd, b[total:])
		if err == unix.EINTR || err == unix.EAGAIN {
			continue
		}
		if err != nil {
			return total, err
		}
		total += n
	}
	// wait until all bytes left the UART (tcdrain)
	if err := unix.IoctlSetInt(p.fd, unix.TCSBRK, 1); err != nil {
		return total, err
	}
	return total, nil
}

func (p *serialPort) Flush() error {
	return unix.IoctlSetInt(p.fd, unix.TCFLSH, unix.TCIFLUSH)
}

func (p *serialPort) Close() error {
	return unix.Close(p.fd)
}