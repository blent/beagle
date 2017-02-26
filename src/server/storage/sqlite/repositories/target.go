package repositories

import (
	"database/sql"
	"fmt"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/server/storage"
	"github.com/blent/beagle/src/server/storage/sqlite/repositories/mappers"
	"github.com/pkg/errors"
)

const (
	selectQuery = "SELECT id, key, name, kind, enabled FROM"
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

func (r *SQLiteTargetRepository) GetById(id uint) (*tracking.Target, error) {
	if id == 0 {
		return nil, errors.New("id must be greater than 0")
	}

	stmt, err := r.db.Prepare(
		fmt.Sprintf(
			"%s %s where id=? LIMIT 1",
			selectQuery,
			r.targetTableName,
		),
	)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return mappers.ToTarget(stmt.QueryRow(id))
}

func (r *SQLiteTargetRepository) GetByKey(key string) (*tracking.Target, error) {
	if key == "" {
		return nil, errors.New("key must be non-empty string")
	}

	stmt, err := r.db.Prepare(
		fmt.Sprintf(
			"%s %s where key='?' LIMIT 1",
			selectQuery,
			r.targetTableName,
		),
	)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return mappers.ToTarget(stmt.QueryRow(key))
}

func (r *SQLiteTargetRepository) Find(query *storage.TargetQuery) ([]*tracking.Target, error) {

	var queryStmt string
	var takeAll bool

	if query == nil || query.Take == 0 {
		takeAll = true
	}

	if !takeAll {
		queryStmt = fmt.Sprintf(
			"%s %s LIMIT ?, ?",
			selectQuery,
			r.targetTableName,
		)
	} else {
		queryStmt = fmt.Sprintf("%s %s", selectQuery, r.targetTableName)
	}

	stmt, err := r.db.Prepare(queryStmt)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()


	var rows *sql.Rows

	if !takeAll {
		rows, err = stmt.Query(query.Take, query.Skip)
	} else {
		rows, err = stmt.Query()
	}

	if err != nil {
		return nil, err
	}

	return mappers.ToTargets(rows, query.Take)
}
