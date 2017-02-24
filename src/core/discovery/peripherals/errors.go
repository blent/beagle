package peripherals

import "github.com/pkg/errors"

var (
	ErrUnsupportedPeripheral = errors.New("unsupported peripheral kind")
	ErrInvalidIBeaconUuid    = errors.New("invalid iBeacon uuid")
)
