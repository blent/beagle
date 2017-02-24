package devices

import (
	"github.com/blent/beagle/src/core/discovery"
	"github.com/blent/beagle/src/core/discovery/peripherals"
	"golang.org/x/net/context"
	"math/rand"
	"strconv"
)

type MockDevice struct {
	isScanning bool
}

func NewMockDevice() *MockDevice {
	return &MockDevice{
		isScanning: false,
	}
}

func (device *MockDevice) IsScanning() bool {
	return device.isScanning
}

func (device *MockDevice) Scan(ctx context.Context) (*discovery.Stream, error) {
	if device.isScanning {
		return nil, ErrStartScanning
	}

	onData := make(chan peripherals.Peripheral, 1000)
	onError := make(chan error)

	device.isScanning = true
	go device.start(ctx, onData, onError)
	go device.stopOnDone(ctx, onData, onError)

	return discovery.NewStream(onData, onError), nil
}

func (device *MockDevice) start(ctx context.Context, inData chan<- peripherals.Peripheral, inError chan<- error) {

}

func (device *MockDevice) stopOnDone(ctx context.Context, inData chan peripherals.Peripheral, inError chan error) {
	<-ctx.Done()
	device.isScanning = false
	close(inData)
	close(inError)
}

func (device *MockDevice) EmitDefaultDiscovery() {
	device.EmitDiscovery(
		strconv.Itoa(rand.Int()),
		peripherals.PERIPHERAL_IBEACON,
		"name",
		[]byte{},
		1,
		1,
		"",
	)
}

func (device *MockDevice) EmitDiscovery(id string, kind string, localName string, data []byte, power float64, rssi float64, address string) {
	//if device.discoveryHandler != nil {
	//	peripheral, err := peripherals.NewMockPeripheral(
	//		id,
	//		kind,
	//		localName,
	//		data,
	//		power,
	//		rssi,
	//		address,
	//	)
	//
	//	if err == nil {
	//		device.discoveryHandler(peripheral)
	//	}
	//}
}
