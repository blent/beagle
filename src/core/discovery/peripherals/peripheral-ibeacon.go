package peripherals

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/go-errors/errors"
	"strconv"
	"strings"
)

var (
	appleCompanyIdentifier        = 0x004c
	iBeaconType                   = 0x02
	expectedIBeaconDataLength     = 0x15
	iBeaconManufacturerDataLength = 25
)

type IBeaconPeripheral struct {
	*GenericPeripheral
	uuid  string
	major uint16
	minor uint16
}

func NewIBeaconPeripheral(localName string, data []byte, power float64, rssi float64, address string) (*IBeaconPeripheral, error) {
	uuid := getIBeaconUuid(data)
	major := getIBeaconMajor(data)
	minor := getIBeaconMinor(data)
	id := CreateIBeaconUniqueKey(uuid, major, minor)

	if id == "" {
		return nil, ErrInvalidIBeaconUuid
	}

	return &IBeaconPeripheral{
		GenericPeripheral: newGenericPeripheral(
			id,
			PERIPHERAL_IBEACON,
			localName,
			data,
			power,
			rssi,
			address,
		),
		uuid:  uuid,
		major: major,
		minor: minor,
	}, nil
}

func (beacon *IBeaconPeripheral) Uuid() string {
	return beacon.uuid
}

func (beacon *IBeaconPeripheral) Major() uint16 {
	return beacon.major
}

func (beacon *IBeaconPeripheral) Minor() uint16 {
	return beacon.minor
}

func CreateIBeaconUniqueKey(uuid string, major uint16, minor uint16) string {
	return fmt.Sprintf(
		"%s:%s:%s",
		uuid,
		strconv.Itoa(int(major)),
		strconv.Itoa(int(minor)),
	)
}

func ParseIBeaconUniqueKey(key string) (string, uint16, uint16, error) {
	arr := strings.Split(key, ":")

	if len(arr) != 3 {
		return "", 0, 0, errors.New("invalid unique key")
	}

	uuid := arr[0]

	major, err := strconv.ParseUint(arr[1], 10, 16)

	if err != nil {
		return "", 0, 0, err
	}

	minor, err := strconv.ParseUint(arr[2], 10, 16)

	if err != nil {
		return "", 0, 0, err
	}

	return uuid, uint16(major), uint16(minor), nil
}

func isIBeacon(data []byte) bool {
	if len(data) < iBeaconManufacturerDataLength {
		return false
	}

	return uint16(data[0]) == uint16(appleCompanyIdentifier) &&
		data[2] == uint8(iBeaconType) &&
		data[3] == uint8(expectedIBeaconDataLength)
}

func getIBeaconUuid(data []byte) string {
	return hex.EncodeToString(data[4:20])
}

func getIBeaconMajor(data []byte) uint16 {
	return binary.BigEndian.Uint16(data[20:22])
}

func getIBeaconMinor(data []byte) uint16 {
	return binary.BigEndian.Uint16(data[22:24])
}
