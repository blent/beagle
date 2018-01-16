package activity

import (
	"github.com/blent/beagle/pkg/discovery/peripherals"
	"github.com/blent/beagle/pkg/notification"
	"go.uber.org/zap"
)

type Writer struct {
	logger *zap.Logger
}

func NewWriter(logger *zap.Logger) *Writer {
	return &Writer{
		logger: logger,
	}
}

func (history *Writer) Use(broker *notification.EventBroker) *Writer {
	if broker == nil {
		return history
	}

	broker.Subscribe(notification.FOUND, func(peripheral peripherals.Peripheral, registered bool) {
		if registered {
			// TODO: Write to DB
		}
	})

	broker.Subscribe(notification.LOST, func(peripheral peripherals.Peripheral, registered bool) {
		if registered {
			// TODO: Write to DB
		}
	})

	return history
}
