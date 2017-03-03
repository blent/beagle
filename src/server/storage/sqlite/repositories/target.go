package repositories

import (
	"database/sql"
	"fmt"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/server/storage"
	"github.com/blent/beagle/src/server/storage/sqlite/repositories/mapping"
	"github.com/pkg/errors"
)

const (
	selectQuery       = "SELECT id, key, name, kind, enabled FROM %s"
	insertQuery       = "INSERT INTO %s (key, name, kind, enabled) VALUES %s"
	insertValuesQuery = "(?, ?, ?, ?)"
	updateQuery       = "UPDATE %s SET name=?, enabled=? WHERE id=?"
	deleteQuery       = "DELETE FROM %s WHERE id=?"
)

type (
	SQLiteTargetRepository struct {
		targetTableName string
		db              *sql.DB
	}
)

func NewSQLiteTargetRepository(targetTableName string, db *sql.DB) *SQLiteTargetRepository {
	return &SQLiteTargetRepository{
		targetTableName,
		db,
	}
}

func (r *SQLiteTargetRepository) GetById(id uint64) (*tracking.Target, error) {
	if id == 0 {
		return nil, errors.New("id must be greater than 0")
	}

	stmt, err := r.db.Prepare(
		fmt.Sprintf(
			"%s WHERE id=? LIMIT 1",
			fmt.Sprintf(
				selectQuery,
				r.targetTableName,
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
			"%s WHERE key='?' LIMIT 1",
			fmt.Sprintf(
				selectQuery,
				r.targetTableName,
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

	orderedSelectQuery := selectQuery + " ORDER BY id"

	if !takeAll {
		queryStmt = fmt.Sprintf(
			"%s LIMIT ? OFFSET ?",
			fmt.Sprintf(
				orderedSelectQuery,
				r.targetTableName,
			),
		)
	} else {
		queryStmt = fmt.Sprintf(orderedSelectQuery, r.targetTableName)
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

func (r *SQLiteTargetRepository) Create(target *tracking.Target) (int64, error) {
	if target == nil {
		return -1, errors.New("target missed")
	}

	var id int64
	var err error

	if target.Id > 0 {
		return -1, errors.New("target already created")
	}

	tx, err := r.db.Begin()

	if err != nil {
		return -1, err
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(insertQuery, r.targetTableName, insertValuesQuery),
	)

	if err != nil {
		return -1, r.rollback(tx, err)
	}

	res, err := stmt.Exec(target.Key, target.Name, target.Kind, r.isEnabled(target))

	if err != nil {
		return -1, r.rollback(tx, err)
	}

	id, err = res.LastInsertId()

	if err != nil {
		return -1, r.rollback(tx, err)
	}

	err = tx.Commit()

	if err != nil {
		return -1, err
	}

	return id, err
}

func (r *SQLiteTargetRepository) Update(target *tracking.Target) error {
	if target == nil {
		return errors.New("target missed")
	}

	var err error

	if target.Id == 0 || target.Id < 0 {
		return errors.New("target not created yet")
	}

	tx, err := r.db.Begin()

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(updateQuery, r.targetTableName),
	)

	if err != nil {
		return r.rollback(tx, err)
	}

	_, err = stmt.Exec(target.Name, r.isEnabled(target), target.Id)

	if err != nil {
		return r.rollback(tx, err)
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}

func (r *SQLiteTargetRepository) Delete(id uint64) error {
	if id == 0 {
		return errors.New("id must be greater than 0")
	}

	var err error

	tx, err := r.db.Begin()

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(deleteQuery, r.targetTableName),
	)

	if err != nil {
		return r.rollback(tx, err)
	}

	_, err = stmt.Exec(id)

	if err != nil {
		return r.rollback(tx, err)
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}

func (r *SQLiteTargetRepository) rollback(tx *sql.Tx, reason error) error {
	rollbackErr := tx.Rollback()

	if rollbackErr != nil {
		return fmt.Errorf("%s:%s", rollbackErr.Error(), reason.Error())
	}

	return reason
}

func (r *SQLiteTargetRepository) isEnabled(target *tracking.Target) int {
	enabled := 0

	if target.Enabled {
		enabled = 1
	}

	return enabled
}
