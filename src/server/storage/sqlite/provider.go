package sqlite

import (
	"database/sql"
	"github.com/blent/beagle/src/server/storage"
	"github.com/blent/beagle/src/server/storage/sqlite/repositories"
)

type SQLiteProvider struct {
	db *sql.DB
}

func NewSQLiteProvider(db *sql.DB) *SQLiteProvider {
	return &SQLiteProvider{
		db,
	}
}

func (provider *SQLiteProvider) GetInitializer() storage.Initializer {
	return initialize
}

func (provider *SQLiteProvider) GetTargetRepository() storage.TargetRepository {
	return repositories.NewSQLiteTargetRepository(
		targetTableName,
		targetSubscriberTableName,
		provider.db,
	)
}

func (provider *SQLiteProvider) GetSubscriberRepository() storage.SubscriberRepository {
	return repositories.NewSQLiteSubscriberRepository(subscriberTableName, provider.db)
}
