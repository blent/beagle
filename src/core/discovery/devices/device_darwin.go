package devices

import (
	"github.com/currantlabs/ble/darwin"
	"go.uber.org/zap"
)

func NewDevice(logger *zap.Logger) (*BleDevice, error) {
	engine, err := darwin.NewDevice()

	if err != nil {
		return nil, err
	}

	return NewBleDevice(logger, engine), nil
}
