package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type (
	Initializer func(tx *sql.Tx) (bool, error)

	Provider interface {
		GetInitializer() Initializer
		GetTargetRepository() TargetRepository
		GetSubscriberRepository() SubscriberRepository
	}
)
