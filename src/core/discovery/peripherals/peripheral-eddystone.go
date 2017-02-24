package peripherals

const (
	EDDYSTONE_VARIANT_URL = "url"
	EDDYSTONE_VARIANT_TLM = "tlm"
	EDDYSTONE_VARIANT_UID = "uid"
)

type (
	EddystonePeripheral struct {
		*GenericPeripheral
		variant string
	}
)

func NewEddystonePeripheral(localName string, data []byte, power float64, rssi float64, address string) (*EddystonePeripheral, error) {
	return &EddystonePeripheral{
		GenericPeripheral: newGenericPeripheral(
			GenerateEddystoneId(),
			PERIPHERAL_EDDYSTONE,
			localName,
			data,
			power,
			rssi,
			address,
		),
		variant: getVariant(),
	}, nil
}

// TODO: add Eddystone support
func isEddystone() bool {
	//v := uuid[0]
	//if len(uuid) > 0 && strings.ToUpper(uuid[0].String()) == eddystoneUUID {
	//	return true
	//}

	return false
}

func GenerateEddystoneId() string {
	return "eddystone"
}

func getVariant() string {
	return EDDYSTONE_VARIANT_URL
}
