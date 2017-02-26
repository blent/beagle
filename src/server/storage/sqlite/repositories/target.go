package repositories

import (
	"database/sql"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/server/storage"
)

type (
	SQLiteTargetRepository struct {
		targetTableName           string
		targetSubscriberTableName string
		db                        *sql.DB
	}
)

func NewSQLiteTargetRepository(targetTableName, targetSubscriberTableName string, db *sql.DB) *SQLiteTargetRepository {
	return &SQLiteTargetRepository{
		targetTableName,
		targetSubscriberTableName,
		db,
	}
}

func (repo *SQLiteTargetRepository) GetByKey(key string) (*tracking.Target, error) {
	return nil, nil
}

func (repo *SQLiteTargetRepository) FindByQuery(query *storage.TargetQuery) ([]*tracking.Target, error) {
	return nil, nil
}
