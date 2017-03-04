package sqlite

import (
	"database/sql"
	"github.com/blent/beagle/src/server/storage"
	"github.com/blent/beagle/src/server/storage/providers/sqlite/repositories"
	"github.com/blent/beagle/src/server/utils"
	_ "github.com/mattn/go-sqlite3"
	"path/filepath"
)

type SQLiteProvider struct {
	db *sql.DB
}

func NewSQLiteProvider(connectionString string) (*SQLiteProvider, error) {
	err := utils.EnsureDirectory(filepath.Dir(connectionString))

	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", connectionString)

	if err != nil {
		return nil, err
	}

	return &SQLiteProvider{
		db,
	}, nil
}

func (provider *SQLiteProvider) GetConnection() *sql.DB {
	return provider.db
}

func (provider *SQLiteProvider) GetInitializer() storage.Initializer {
	return initialize
}

func (provider *SQLiteProvider) GetTargetRepository() storage.TargetRepository {
	return repositories.NewSQLiteTargetRepository(
		targetTableName,
		provider.db,
	)
}

func (provider *SQLiteProvider) GetEndpointRepository() storage.EndpointRepository {
	return repositories.NewSQLiteEndpointRepository(
		endpointTableName,
		provider.db,
	)
}

func (provider *SQLiteProvider) GetSubscriberRepository() storage.SubscriberRepository {
	return repositories.NewSQLiteSubscriberRepository(
		subscriberTableName,
		endpointTableName,
		provider.db,
	)
}

func (provider *SQLiteProvider) Close() error {
	return provider.db.Close()
}
