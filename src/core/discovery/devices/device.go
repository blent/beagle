package devices

import (
	"github.com/blent/beagle/src/core/discovery"
	"golang.org/x/net/context"
)

type (
	Device interface {
		IsScanning() bool
		Scan(context.Context) (*discovery.Stream, error)
	}
)
