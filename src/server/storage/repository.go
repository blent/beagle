package storage

import "github.com/blent/beagle/src/core/tracking"

type (
	Query struct {
		Take uint64
		Skip uint64
	}

	TargetQuery struct {
		*Query
		Status string
	}

	TargetRepository interface {
		GetById(uint64) (*tracking.Target, error)
		GetByKey(string) (*tracking.Target, error)
		Find(*TargetQuery) ([]*tracking.Target, error)
		Create(*tracking.Target) (int64, error)
	}

	SubscriberRepository interface {
		GetById(uint64) (*tracking.Subscriber, error)
		GetByName(string) (*tracking.Subscriber, error)
	}

	ActivityHistoryRepository interface{}

	DeliveryHistoryRepository interface{}
)

func NewQuery(take, skip uint64) *Query {
	return &Query{take, skip}
}

func NewTargetQuery(take, skip uint64, status string) *TargetQuery {
	return &TargetQuery{
		Query:  NewQuery(take, skip),
		Status: status,
	}
}
