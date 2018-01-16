package delivery

import "github.com/pkg/errors"

var (
	ErrUnsupportedEventName        = errors.New("unsupported event name")
	ErrUnsupportedHttpMethod       = errors.New("unsupported http method")
	ErrUnableToSerializePeripheral = errors.New("unable to serialize peripheral")
)
