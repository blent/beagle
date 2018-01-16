package activity

import (
	"github.com/blent/beagle/pkg/notification"
	"github.com/bradfitz/slice"
	"go.uber.org/zap"
	"sync"
)

type Monitoring struct {
	mu      *sync.RWMutex
	logger  *zap.Logger
	records map[string]*Record
}

func New(logger *zap.Logger) *Monitoring {
	return &Monitoring{
		mu:      &sync.RWMutex{},
		logger:  logger,
		records: make(map[string]*Record),
	}
}

func (s *Monitoring) Quantity() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.records)
}

func (s *Monitoring) GetRecords(take, skip int) []*Record {
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

func (s *Monitoring) Use(broker *notification.Broker) *Monitoring {
	if broker == nil {
		return s
	}

	broker.AddEventListener(func(evt notification.Event) {
		s.mu.Lock()
		defer s.mu.Unlock()

		peripheral := evt.Peripheral

		if evt.Name == notification.FOUND {
			s.records[peripheral.UniqueKey()] = &Record{
				Key:        peripheral.UniqueKey(),
				Kind:       peripheral.Kind(),
				Proximity:  peripheral.Proximity(),
				Registered: evt.Registered,
				Time:       evt.Timestamp,
			}
		} else {
			delete(s.records, peripheral.UniqueKey())
		}
	})

	return s
}
