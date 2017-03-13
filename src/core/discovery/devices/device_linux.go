package devices

import (
	"github.com/blent/beagle/src/core/logging"
	"github.com/currantlabs/ble/linux"
)

func NewDevice(logger *logging.Logger) (*BleDevice, error) {
	engine, err := linux.NewDevice()

	if err != nil {
		return nil, err
	}

	return NewBleDevice(logger, engine), nil
}
