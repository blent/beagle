package activity

import (
	"github.com/blent/beagle/src/core/discovery/peripherals"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/notification"
)

type Writer struct {
	logger *logging.Logger
}

func NewWriter(logger *logging.Logger) *Writer {
	return &Writer{
		logger: logger,
	}
}

func (history *Writer) Use(broker *notification.EventBroker) *Writer {
	if broker == nil {
		return history
	}

	broker.Subscribe(notification.PERIPHERAL_FOUND, func(peripheral peripherals.Peripheral, registered bool) {
		if registered {
			// TODO: Write to DB
		}
	})

	broker.Subscribe(notification.PERIPHERAL_LOST, func(peripheral peripherals.Peripheral, registered bool) {
		if registered {
			// TODO: Write to DB
		}
	})

	return history
}
