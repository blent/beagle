package devices

import (
	"github.com/blent/beagle/src/core/discovery"
	"github.com/blent/beagle/src/core/discovery/peripherals"
	"github.com/currantlabs/ble"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type (
	Device interface {
		IsScanning() bool
		Scan(context.Context) (*discovery.Stream, error)
	}

	BleDevice struct {
		isScanning bool
		logger     *zap.Logger
		engine     ble.Device
	}
)

const bufferSize = 1000

func NewBleDevice(logger *zap.Logger, engine ble.Device) *BleDevice {
	ble.SetDefaultDevice(engine)

	device := &BleDevice{
		isScanning: false,
		logger:     logger,
		engine:     engine,
	}

	return device
}

func (device *BleDevice) IsScanning() bool {
	return device.isScanning
}

func (device *BleDevice) Scan(ctx context.Context) (*discovery.Stream, error) {
	if device.isScanning {
		return nil, ErrStartScanning
	}

	onData := make(chan peripherals.Peripheral, bufferSize)
	onError := make(chan error)

	device.isScanning = true
	go device.start(ctx, onData, onError)
	go device.stopOnDone(ctx, onData, onError)

	return discovery.NewStream(onData, onError), nil
}

func (device *BleDevice) start(ctx context.Context, inData chan<- peripherals.Peripheral, inError chan<- error) {
	err := ble.Scan(ctx, true, func(adv ble.Advertisement) {
		localName := adv.LocalName()
		manufacturerData := adv.ManufacturerData()

		if !peripherals.IsSupportedPeripheral(manufacturerData) {
			return
		}

		peripheral, err := peripherals.NewPeripheral(
			localName,
			manufacturerData,
			float64(adv.TxPowerLevel()),
			float64(adv.RSSI()),
			adv.Address().String(),
		)

		if err == nil {
			inData <- peripheral
		} else {
			device.logger.Error(
				"failed to parse peripheral",
				zap.Error(err),
			)
		}
	}, nil)

	if err != nil {
		device.isScanning = false
		inError <- err
	}
}

func (device *BleDevice) stopOnDone(ctx context.Context, inData chan peripherals.Peripheral, inError chan error) {
	<-ctx.Done()
	device.isScanning = false
	close(inData)
	close(inError)
}
