package peripherals

type (
	MockPeripheral struct {
		*GenericPeripheral
	}
)

func NewMockPeripheral(id string, kind string, localName string, data []byte, power float64, rssi float64, address string) *MockPeripheral {
	return &MockPeripheral{
		GenericPeripheral: newGenericPeripheral(
			id,
			kind,
			localName,
			data,
			power,
			rssi,
			address,
		),
	}
}
