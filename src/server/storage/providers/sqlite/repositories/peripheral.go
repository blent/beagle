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
	peripheralSelectQuery       = "SELECT id, key, name, kind, enabled FROM %s"
	peripheralInsertQuery       = "INSERT INTO %s (key, name, kind, enabled) VALUES %s"
	peripheralInsertValuesQuery = "(?, ?, ?, ?)"
	peripheralUpdateQuery       = "UPDATE %s SET name=?, enabled=? WHERE id=?"
	peripheralDeleteQuery       = "DELETE FROM %s WHERE id=?"
)

type (
	SQLitePeripheralRepository struct {
		tableName string
		db        *sql.DB
	}
)

func NewSQLitePeripheralRepository(tableName string, db *sql.DB) *SQLitePeripheralRepository {
	return &SQLitePeripheralRepository{
		tableName,
		db,
	}
}

func (r *SQLitePeripheralRepository) Get(id uint64) (*tracking.Peripheral, error) {
	if id == 0 {
		return nil, errors.New("id must be greater than 0")
	}

	stmt, err := r.db.Prepare(
		fmt.Sprintf(
			"%s WHERE id=? LIMIT 1",
			fmt.Sprintf(
				peripheralSelectQuery,
				r.tableName,
			),
		),
	)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return mapping.ToPeripheral(stmt.QueryRow(id))
}

func (r *SQLitePeripheralRepository) GetByKey(key string) (*tracking.Peripheral, error) {
	if key == "" {
		return nil, errors.New("key must be non-empty string")
	}

	stmt, err := r.db.Prepare(
		fmt.Sprintf(
			"%s WHERE key=? LIMIT 1",
			fmt.Sprintf(
				peripheralSelectQuery,
				r.tableName,
			),
		),
	)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return mapping.ToPeripheral(stmt.QueryRow(key))
}

func (r *SQLitePeripheralRepository) Find(query *storage.PeripheralQuery) ([]*tracking.Peripheral, error) {
	var queryStmt string
	var takeAll bool

	if query == nil || query.Take == 0 {
		takeAll = true
	}

	orderedSelectQuery := peripheralSelectQuery + " ORDER BY id"

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

	return mapping.ToPeripherals(rows, query.Take)
}

func (r *SQLitePeripheralRepository) Create(target *tracking.Peripheral, tx *sql.Tx) (uint64, error) {
	if target == nil {
		return 0, errors.New("peripheral missed")
	}

	var id int64

	if target.Id > 0 {
		return 0, errors.New("peripheral already created")
	}

	tx, closeTx, err := storage.TryToBegin(r.db, tx)

	if err != nil {
		return 0, err
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(peripheralInsertQuery, r.tableName, peripheralInsertValuesQuery),
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

func (r *SQLitePeripheralRepository) Update(target *tracking.Peripheral, tx *sql.Tx) error {
	if target == nil {
		return errors.New("peripheral missed")
	}

	if target.Id == 0 || target.Id < 0 {
		return errors.New("peripheral not created yet")
	}

	tx, closeTx, err := storage.TryToBegin(r.db, tx)

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(peripheralUpdateQuery, r.tableName),
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

func (r *SQLitePeripheralRepository) Delete(id uint64, tx *sql.Tx) error {
	if id == 0 {
		return errors.New("id must be greater than 0")
	}

	tx, closeTx, err := storage.TryToBegin(r.db, tx)

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(peripheralDeleteQuery, r.tableName),
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
