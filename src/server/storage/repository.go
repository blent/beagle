package storage

import "github.com/blent/beagle/src/core/tracking"

type (
	Query struct {
		Take int64
		Skip int64
	}

	TargetQuery struct {
		*Query
		Status uint64
	}

	TargetRepository interface {
		GetByKey(string) (*tracking.Target, error)
	}

	SubscriberRepository interface {
		GetByName(string) (*tracking.Subscriber, error)
	}

	ActivityHistoryRepository interface{}

	DeliveryHistoryRepository interface{}
)
