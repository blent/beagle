package activity

import (
	"github.com/blent/beagle/src/core/discovery/peripherals"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/notification"
	"github.com/blent/beagle/src/core/tracking"
	"sync"
	"time"
)

type Writer struct {
	mu      sync.Mutex
	logger  *logging.Logger
	records map[string]*Record
}

func NewWriter(logger *logging.Logger) *Writer {
	return &Writer{
		logger:  logger,
		records: make(map[string]*Record),
	}
}

func (history *Writer) GetCurrent() []*Record {
	history.mu.Lock()
	defer history.mu.Unlock()

	list := make([]*Record, 0, len(history.records))

	for _, record := range history.records {
		var item *Record

		// copying..
		*item = *record
		list = append(list, item)
	}

	return list
}

func (history *Writer) Use(broker *notification.EventBroker) *Writer {
	if broker == nil {
		return history
	}

	broker.Subscribe(notification.PERIPHERAL_FOUND, func(target *tracking.Target, peripheral peripherals.Peripheral) {
		history.mu.Lock()
		defer history.mu.Unlock()

		history.records[target.Key] = &Record{
			Key:       target.Key,
			Kind:      target.Kind,
			Proximity: peripheral.Proximity(),
			Time:      time.Now(),
		}
	})

	broker.Subscribe(notification.PERIPHERAL_LOST, func(target *tracking.Target, peripheral peripherals.Peripheral) {
		history.mu.Lock()
		defer history.mu.Unlock()

		delete(history.records, target.Key)
	})

	return history
}
