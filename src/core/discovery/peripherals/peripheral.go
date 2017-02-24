package peripherals

import (
	"math"
)

type (
	Peripheral interface {
		UniqueKey() string

		LocalName() string

		Kind() string

		ManufacturerData() []byte

		TxPowerLevel() float64

		RSSI() float64

		Address() string

		Proximity() string

		Accuracy() float64
	}

	GenericPeripheral struct {
		uniqueKey        string
		localName        string
		kind             string
		manufacturerData []byte
		txPowerLevel     float64
		rssi             float64
		address          string
		proximity        string
		accuracy         float64
	}
)

func (peripheral *GenericPeripheral) UniqueKey() string {
	return peripheral.uniqueKey
}

func (peripheral *GenericPeripheral) LocalName() string {
	return peripheral.localName
}

func (peripheral *GenericPeripheral) Kind() string {
	return peripheral.kind
}

func (peripheral *GenericPeripheral) ManufacturerData() []byte {
	return peripheral.manufacturerData
}

func (peripheral *GenericPeripheral) TxPowerLevel() float64 {
	return peripheral.txPowerLevel
}

func (peripheral *GenericPeripheral) RSSI() float64 {
	return peripheral.rssi
}

func (peripheral *GenericPeripheral) Address() string {
	return peripheral.address
}

func (peripheral *GenericPeripheral) Proximity() string {
	return peripheral.proximity
}

func (peripheral *GenericPeripheral) Accuracy() float64 {
	return peripheral.accuracy
}

func NewPeripheral(localName string, data []byte, power float64, rssi float64, address string) (Peripheral, error) {
	if isIBeacon(data) {
		return NewIBeaconPeripheral(localName, data, power, rssi, address)
	}

	return nil, ErrUnsupportedPeripheral
}

func IsSupportedPeripheral(data []byte) bool {
	return isIBeacon(data) || isEddystone()
}

func newGenericPeripheral(uniqueKey string, kind string, localName string, data []byte, power float64, rssi float64, address string) *GenericPeripheral {
	accuracy := calculateAccuracy(power, rssi)

	return &GenericPeripheral{
		uniqueKey:        uniqueKey,
		localName:        localName,
		kind:             kind,
		manufacturerData: data,
		txPowerLevel:     power,
		rssi:             rssi,
		address:          address,
		proximity:        calculateProximity(accuracy),
		accuracy:         accuracy,
	}
}

func calculateAccuracy(power float64, rssi float64) float64 {
	return math.Pow(12.0, 1.5*((rssi/power)-1))
}

func calculateProximity(accuracy float64) string {
	if accuracy < 0 {
		return PROXIMITY_UKNOWN
	} else if accuracy < 0.5 {
		return PROXIMITY_IMMEDIATE
	} else if accuracy < 4.0 {
		return PROXIMITY_NEAR
	}

	return PROXIMITY_FAR
}
