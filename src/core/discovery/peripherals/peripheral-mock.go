package peripherals

type (
	MockPeripheral struct {
		*GenericPeripheral
		variant string
	}
)

func NewMockPeripheral(id string, kind string, localName string, data []byte, power float64, rssi float64, address string) (*EddystonePeripheral, error) {
	return &EddystonePeripheral{
		GenericPeripheral: newGenericPeripheral(
			id,
			kind,
			localName,
			data,
			power,
			rssi,
			address,
		),
	}, nil
}
