//go:build !linux

package modbus

import "errors"

// OpenSerial is only implemented on Linux. On other platforms use
// device "fake" in the config (simulated charge controller).
func OpenSerial(device string, baud int) (Conn, error) {
	return nil, errors.New("serial port access is only supported on linux")
}