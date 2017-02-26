package devices

import (
	"github.com/blent/beagle/src/core/discovery"
	"github.com/blent/beagle/src/core/discovery/peripherals"
	"github.com/blent/beagle/src/core/logging"
	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/linux"
	"golang.org/x/net/context"
)

const bufferSize = 1000

type BleDevice struct {
	isScanning bool
	logger     *logging.Logger
	engine     ble.Device
}

func NewBleDevice(logger *logging.Logger) (*BleDevice, error) {
	engine, err := linux.NewDevice()

	if err != nil {
		return nil, err
	}

	ble.SetDefaultDevice(engine)

	device := &BleDevice{
		isScanning: false,
		logger:     logger,
		engine:     engine,
	}

	return device, nil
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
			device.logger.Errorf("failed to parse peripheral: %s", err.Error())
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
