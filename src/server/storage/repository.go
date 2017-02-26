package storage

import "github.com/blent/beagle/src/core/tracking"

type (
	Query struct {
		Take uint
		Skip uint
	}

	TargetQuery struct {
		*Query
		Status string
	}

	TargetRepository interface {
		GetById(uint) (*tracking.Target, error)
		GetByKey(string) (*tracking.Target, error)
		Find(*TargetQuery) ([]*tracking.Target, error)
	}

	SubscriberRepository interface {
		GetById(uint) (*tracking.Subscriber, error)
		GetByName(string) (*tracking.Subscriber, error)
	}

	ActivityHistoryRepository interface{}

	DeliveryHistoryRepository interface{}
)

func NewQuery(take, skip uint) *Query {
	return &Query{take, skip }
}

func NewTargetQuery(take, skip uint, status string) *TargetQuery {
	return &TargetQuery{
		Query: NewQuery(take, skip),
		Status: status,
	}
}
