package delivery

import (
	"github.com/blent/beagle/pkg/delivery"
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

func (history *Writer) Use(sender *delivery.Sender) {
	if sender == nil {
		return
	}

	sender.AddEventListener(func(evt delivery.Event) {

	})
}
