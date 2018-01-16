package storage

import (
	"database/sql"
	"github.com/blent/beagle/pkg/notification"
	"github.com/blent/beagle/pkg/tracking"
)

type (
	DeletionQuery struct {
		Id      []uint64
		InRange bool
	}

	Pagination struct {
		Take uint64
		Skip uint64
	}

	PeripheralFilter struct {
		Status string
	}

	PeripheralQuery struct {
		*Pagination
		*PeripheralFilter
	}

	EndpointFilter struct {
		Name string
	}

	EndpointQuery struct {
		*Pagination
		*EndpointFilter
	}

	SubscriberFilter struct {
		TargetId uint64
		Events   []string
		Status   string
	}

	SubscriberQuery struct {
		*Pagination
		*SubscriberFilter
	}

	PeripheralRepository interface {
		Find(*PeripheralQuery) ([]*tracking.Peripheral, error)
		Count(*PeripheralFilter) (uint64, error)
		GetByKey(string) (*tracking.Peripheral, error)
		Get(uint64) (*tracking.Peripheral, error)
		Create(*tracking.Peripheral, *sql.Tx) (uint64, error)
		Update(*tracking.Peripheral, *sql.Tx) error
		Delete(uint64, *sql.Tx) error
		DeleteMany(*DeletionQuery, *sql.Tx) error
	}

	SubscriberRepository interface {
		Find(*SubscriberQuery) ([]*notification.Subscriber, error)
		Count(*SubscriberFilter) (uint64, error)
		Get(uint64) (*notification.Subscriber, error)
		Create(*notification.Subscriber, uint64, *sql.Tx) (uint64, error)
		CreateMany([]*notification.Subscriber, uint64, *sql.Tx) error
		Update(*notification.Subscriber, *sql.Tx) error
		UpdateMany([]*notification.Subscriber, *sql.Tx) error
		Delete(uint64, *sql.Tx) error
		DeleteMany(*DeletionQuery, *sql.Tx) error
	}

	EndpointRepository interface {
		Get(uint64) (*notification.Endpoint, error)
		Count(*EndpointFilter) (uint64, error)
		Find(*EndpointQuery) ([]*notification.Endpoint, error)
		Create(*notification.Endpoint, *sql.Tx) (uint64, error)
		Update(*notification.Endpoint, *sql.Tx) error
		Delete(uint64, *sql.Tx) error
		DeleteMany(*DeletionQuery, *sql.Tx) error
	}

	ActivityHistoryRepository interface{}

	DeliveryHistoryRepository interface{}
)

func NewPagination(take, skip uint64) *Pagination {
	return &Pagination{take, skip}
}

func NewTargetQuery(take, skip uint64, status string) *PeripheralQuery {
	return &PeripheralQuery{
		Pagination: NewPagination(take, skip),
		PeripheralFilter: &PeripheralFilter{
			status,
		},
	}
}

func NewEndpointQuery(take, skip uint64, name string) *EndpointQuery {
	return &EndpointQuery{
		Pagination: NewPagination(take, skip),
		EndpointFilter: &EndpointFilter{
			Name: name,
		},
	}
}

func NewSubscriberQuery(take, skip, targetId uint64, events []string, status string) *SubscriberQuery {
	return &SubscriberQuery{
		Pagination: NewPagination(take, skip),
		SubscriberFilter: &SubscriberFilter{
			targetId,
			events,
			status,
		},
	}
}
