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

func (s *Service) Quantity() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.records)
}

func (s *Service) GetRecords(take, skip int) []*Record {
	s.mu.Lock()
	defer s.mu.Unlock()

	resultSize := take

	if take == 0 {
		resultSize = len(s.records)
	}

	list := make([]*Record, 0, len(s.records))
	result := make([]*Record, 0, resultSize)

	for _, record := range s.records {
		list = append(list, record)
	}

	for idx, record := range list {
		num := idx + 1
		if skip == 0 || skip > num  {
			if len(result) == resultSize {
				break
			}

			// copying..
			item := *record
			list = append(list, &item)
		}
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
