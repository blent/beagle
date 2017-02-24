package devices

import "github.com/pkg/errors"

var (
	ErrStartScanning = errors.New("device is already started scanning")
	ErrStopScanning  = errors.New("device is already stopped scanning")
)
