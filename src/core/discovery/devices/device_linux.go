package devices

import (
	"github.com/currantlabs/ble/linux"
	"go.uber.org/zap"
)

func NewDevice(logger *zap.Logger) (*BleDevice, error) {
	engine, err := linux.NewDevice()

	if err != nil {
		return nil, err
	}

	return NewBleDevice(logger, engine), nil
}
