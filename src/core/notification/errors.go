package notification

import "github.com/pkg/errors"

var (
	ErrUnsupportedEventName         = errors.New("unsupported event name")
	ErrUnsupportedHttpMethod        = errors.New("unsupported http method")
	ErrUnabledToSerializePeripheral = errors.New("unabled to serialize peripheral")
)
