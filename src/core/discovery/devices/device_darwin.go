package devices

import (
	"github.com/go-ble/ble/darwin"
	"go.uber.org/zap"
)

func NewDevice(logger *zap.Logger) (*BleDevice, error) {
	engine, err := darwin.NewDevice()

	if err != nil {
		return nil, err
	}

	return NewBleDevice(logger, engine), nil
}
