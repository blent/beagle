package discovery

import (
	"github.com/blent/beagle/pkg/discovery/peripherals"
)

type Stream struct {
	data  <-chan peripherals.Peripheral
	error <-chan error
}

func NewStream(data <-chan peripherals.Peripheral, error <-chan error) *Stream {
	return &Stream{
		data:  data,
		error: error,
	}
}

func (stream *Stream) Data() <-chan peripherals.Peripheral {
	return stream.data
}

func (stream *Stream) Error() <-chan error {
	return stream.error
}
