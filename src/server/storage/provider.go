package storage

import (
	"database/sql"
)

type (
	Initializer func(tx *sql.Tx) (bool, error)

	Provider interface {
		GetConnection() *sql.DB
		GetInitializer() Initializer
		GetTargetRepository() TargetRepository
		GetSubscriberRepository() SubscriberRepository
		Close() error
	}
)
