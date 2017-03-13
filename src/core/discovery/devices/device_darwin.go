package devices

import (
	"github.com/blent/beagle/src/core/logging"
	"github.com/currantlabs/ble/darwin"
)

func NewDevice(logger *logging.Logger) (*BleDevice, error) {
	engine, err := darwin.NewDevice()

	if err != nil {
		return nil, err
	}

	return NewBleDevice(logger, engine), nil
}
