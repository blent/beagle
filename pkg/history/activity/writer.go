package activity

import (
	"github.com/blent/beagle/pkg/notification"
	"go.uber.org/zap"
)

type Writer struct {
	logger *zap.Logger
}

func New(logger *zap.Logger) *Writer {
	return &Writer{
		logger: logger,
	}
}

func (history *Writer) Use(broker *notification.Broker) {
	if broker == nil {
		return
	}

	broker.AddEventListener(func(evt notification.Event) {

	})
}
