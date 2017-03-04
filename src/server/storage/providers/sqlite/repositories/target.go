package repositories

import (
	"database/sql"
	"fmt"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/server/storage"
	"github.com/blent/beagle/src/server/storage/providers/sqlite/repositories/mapping"
	"github.com/pkg/errors"
)

const (
	targetSelectQuery       = "SELECT id, key, name, kind, enabled FROM %s"
	targetInsertQuery       = "INSERT INTO %s (key, name, kind, enabled) VALUES %s"
	targetInsertValuesQuery = "(?, ?, ?, ?)"
	targetUpdateQuery       = "UPDATE %s SET name=?, enabled=? WHERE id=?"
	targetDeleteQuery       = "DELETE FROM %s WHERE id=?"
)

type (
	SQLiteTargetRepository struct {
		tableName string
		db        *sql.DB
	}
)

func NewSQLiteTargetRepository(tableName string, db *sql.DB) *SQLiteTargetRepository {
	return &SQLiteTargetRepository{
		tableName,
		db,
	}
}

func (r *SQLiteTargetRepository) Get(id uint64) (*tracking.Target, error) {
	if id == 0 {
		return nil, errors.New("id must be greater than 0")
	}

	stmt, err := r.db.Prepare(
		fmt.Sprintf(
			"%s WHERE id=? LIMIT 1",
			fmt.Sprintf(
				targetSelectQuery,
				r.tableName,
			),
		),
	)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return mapping.ToTarget(stmt.QueryRow(id))
}

func (r *SQLiteTargetRepository) GetByKey(key string) (*tracking.Target, error) {
	if key == "" {
		return nil, errors.New("key must be non-empty string")
	}

	stmt, err := r.db.Prepare(
		fmt.Sprintf(
			"%s WHERE key=? LIMIT 1",
			fmt.Sprintf(
				targetSelectQuery,
				r.tableName,
			),
		),
	)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return mapping.ToTarget(stmt.QueryRow(key))
}

func (r *SQLiteTargetRepository) Find(query *storage.TargetQuery) ([]*tracking.Target, error) {
	var queryStmt string
	var takeAll bool

	if query == nil || query.Take == 0 {
		takeAll = true
	}

	orderedSelectQuery := targetSelectQuery + " ORDER BY id"

	if !takeAll {
		queryStmt = fmt.Sprintf(
			"%s LIMIT ? OFFSET ?",
			fmt.Sprintf(
				orderedSelectQuery,
				r.tableName,
			),
		)
	} else {
		queryStmt = fmt.Sprintf(orderedSelectQuery, r.tableName)
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

	return mapping.ToTargets(rows, query.Take)
}

func (r *SQLiteTargetRepository) Create(target *tracking.Target, tx *sql.Tx) (uint64, error) {
	if target == nil {
		return 0, errors.New("target missed")
	}

	var id int64

	if target.Id > 0 {
		return 0, errors.New("target already created")
	}

	tx, closeTx, err := storage.TryToBegin(r.db, tx)

	if err != nil {
		return 0, err
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(targetInsertQuery, r.tableName, targetInsertValuesQuery),
	)

	if err != nil {
		return 0, storage.TryToRollback(tx, err, closeTx)
	}

	res, err := stmt.Exec(target.Key, target.Name, target.Kind, boolToInt(target.Enabled))

	if err != nil {
		return 0, storage.TryToRollback(tx, err, closeTx)
	}

	id, err = res.LastInsertId()

	if err != nil {
		return 0, storage.TryToRollback(tx, err, closeTx)
	}

	err = storage.TryToCommit(tx, closeTx)

	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

func (r *SQLiteTargetRepository) Update(target *tracking.Target, tx *sql.Tx) error {
	if target == nil {
		return errors.New("target missed")
	}

	if target.Id == 0 || target.Id < 0 {
		return errors.New("target not created yet")
	}

	tx, closeTx, err := storage.TryToBegin(r.db, tx)

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(targetUpdateQuery, r.tableName),
	)

	if err != nil {
		return storage.TryToRollback(tx, err, closeTx)
	}

	_, err = stmt.Exec(target.Name, boolToInt(target.Enabled), target.Id)

	if err != nil {
		return storage.TryToRollback(tx, err, closeTx)
	}

	return storage.TryToCommit(tx, closeTx)
}

func (r *SQLiteTargetRepository) Delete(id uint64, tx *sql.Tx) error {
	if id == 0 {
		return errors.New("id must be greater than 0")
	}

	tx, closeTx, err := storage.TryToBegin(r.db, tx)

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(targetDeleteQuery, r.tableName),
	)

	if err != nil {
		return storage.TryToRollback(tx, err, closeTx)
	}

	_, err = stmt.Exec(id)

	if err != nil {
		return storage.TryToRollback(tx, err, closeTx)
	}

	return storage.TryToCommit(tx, closeTx)
}

func boolToInt(val bool) int {
	enabled := 0

	if val {
		enabled = 1
	}

	return enabled
}
