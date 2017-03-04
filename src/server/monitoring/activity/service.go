package activity

import (
	"github.com/blent/beagle/src/core/discovery/peripherals"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/notification"
	"sync"
	"time"
)

type Service struct {
	mu      sync.Mutex
	logger  *logging.Logger
	records map[string]*Record
}

func NewService(logger *logging.Logger) *Service {
	return &Service{
		logger:  logger,
		records: make(map[string]*Record),
	}
}

func (s *Service) GetRecords() []*Record {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := make([]*Record, 0, len(s.records))

	for _, record := range s.records {
		// copying..
		item := *record
		list = append(list, &item)
	}

	return list
}

func (s *Service) Use(broker *notification.EventBroker) *Service {
	if broker == nil {
		return s
	}

	broker.Subscribe(notification.PERIPHERAL_FOUND, func(peripheral peripherals.Peripheral, registered bool) {
		s.mu.Lock()
		defer s.mu.Unlock()

		s.records[peripheral.UniqueKey()] = &Record{
			Key:        peripheral.UniqueKey(),
			Kind:       peripheral.Kind(),
			Proximity:  peripheral.Proximity(),
			Registered: registered,
			Time:       time.Now(),
		}
	})

	broker.Subscribe(notification.PERIPHERAL_LOST, func(peripheral peripherals.Peripheral, registered bool) {
		s.mu.Lock()
		defer s.mu.Unlock()

		delete(s.records, peripheral.UniqueKey())
	})

	return s
}
