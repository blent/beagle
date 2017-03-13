package storage

import (
	"database/sql"
)

type (
	Initializer func(tx *sql.Tx) (bool, error)

	Provider interface {
		GetConnection() *sql.DB
		GetInitializer() Initializer
		GetPeripheralRepository() PeripheralRepository
		GetSubscriberRepository() SubscriberRepository
		GetEndpointRepository() EndpointRepository
		Close() error
	}
)
