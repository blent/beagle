package tracking

import "github.com/blent/beagle/pkg/discovery/peripherals"

type Stream struct {
	found <-chan peripherals.Peripheral
	lost  <-chan peripherals.Peripheral
	error <-chan error
}

func NewStream(found <-chan peripherals.Peripheral, lost <-chan peripherals.Peripheral, error <-chan error) *Stream {
	return &Stream{found, lost, error}
}

func (stream *Stream) Found() <-chan peripherals.Peripheral {
	return stream.found
}

func (stream *Stream) Lost() <-chan peripherals.Peripheral {
	return stream.lost
}

func (stream *Stream) Error() <-chan error {
	return stream.error
}
