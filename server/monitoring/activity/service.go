package activity

import (
	"github.com/blent/beagle/pkg/discovery/peripherals"
	"github.com/blent/beagle/pkg/notification"
	"github.com/bradfitz/slice"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Service struct {
	mu      *sync.RWMutex
	logger  *zap.Logger
	records map[string]*Record
}

func NewService(logger *zap.Logger) *Service {
	return &Service{
		mu:      &sync.RWMutex{},
		logger:  logger,
		records: make(map[string]*Record),
	}
}

func (s *Service) Quantity() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.records)
}

func (s *Service) GetRecords(take, skip int) []*Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resultSize := take

	if take == 0 {
		resultSize = len(s.records)
	}

	// convert map to list
	list := make([]*Record, 0, len(s.records))
	result := make([]*Record, 0, resultSize)

	for _, record := range s.records {
		list = append(list, record)
	}

	slice.Sort(list, func(i, j int) bool {
		return list[i].Key > list[j].Key
	})

	for idx, record := range list {
		num := idx + 1
		if skip == 0 || skip > num {
			if len(result) == resultSize {
				break
			}

			// copying..
			item := *record
			result = append(result, &item)
		}
	}

	return result
}

func (s *Service) Use(broker *notification.EventBroker) *Service {
	if broker == nil {
		return s
	}

	broker.Subscribe(notification.FOUND, func(peripheral peripherals.Peripheral, registered bool) {
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

	broker.Subscribe(notification.LOST, func(peripheral peripherals.Peripheral, registered bool) {
		s.mu.Lock()
		defer s.mu.Unlock()

		delete(s.records, peripheral.UniqueKey())
	})

	return s
}
