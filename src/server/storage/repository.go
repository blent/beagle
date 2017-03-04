package storage

import (
	"database/sql"
	"github.com/blent/beagle/src/core/notification"
	"github.com/blent/beagle/src/core/tracking"
)

type (
	Query struct {
		Take uint64
		Skip uint64
	}

	TargetQuery struct {
		*Query
		Status string
	}

	EndpointQuery struct {
		*Query
	}

	SubscriberQuery struct {
		*Query
		TargetId uint64
		Event    string
	}

	TargetRepository interface {
		Find(*TargetQuery) ([]*tracking.Target, error)
		GetByKey(string) (*tracking.Target, error)
		Get(uint64) (*tracking.Target, error)
		Create(*tracking.Target, *sql.Tx) (uint64, error)
		Update(*tracking.Target, *sql.Tx) error
		Delete(uint64, *sql.Tx) error
	}

	SubscriberRepository interface {
		Find(*SubscriberQuery) ([]*notification.Subscriber, error)
		Get(uint64) (*notification.Subscriber, error)
		Create(*notification.Subscriber, uint64, *sql.Tx) (uint64, error)
		CreateMany([]*notification.Subscriber, uint64, *sql.Tx) error
		Update(*notification.Subscriber, *sql.Tx) error
		UpdateMany([]*notification.Subscriber, *sql.Tx) error
		Delete(uint64, *sql.Tx) error
	}

	EndpointRepository interface {
		Get(uint64) (*notification.Endpoint, error)
		Find(*EndpointQuery) ([]*notification.Endpoint, error)
		Create(*notification.Endpoint, *sql.Tx) (uint64, error)
		Update(*notification.Endpoint, *sql.Tx) error
		Delete(uint64, *sql.Tx) error
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

func NewEndpointQuery(take, skip uint64) *EndpointQuery {
	return &EndpointQuery{
		Query: NewQuery(take, skip),
	}
}

func NewSubscriberQuery(take, skip, targetId uint64, event string) *SubscriberQuery {
	return &SubscriberQuery{
		Query:    NewQuery(take, skip),
		TargetId: targetId,
		Event:    event,
	}
}
