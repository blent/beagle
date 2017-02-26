package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func NewDatabase(settings *Settings) (*sql.DB, error) {
	return sql.Open(settings.Provider, settings.ConnectionString)
}
