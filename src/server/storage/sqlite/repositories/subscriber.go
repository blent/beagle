package repositories

import (
	"database/sql"
	"github.com/blent/beagle/src/core/tracking"
)

type SQLiteSubscriberRepository struct {
	tableName string
	db        *sql.DB
}

func NewSQLiteSubscriberRepository(tableName string, db *sql.DB) *SQLiteSubscriberRepository {
	return &SQLiteSubscriberRepository{tableName, db}
}

func (r *SQLiteSubscriberRepository) GetById(id uint) (*tracking.Subscriber, error) {
	return nil, nil
}

func (r *SQLiteSubscriberRepository) GetByName(string) (*tracking.Subscriber, error) {
	return nil, nil
}
