package repositories

import (
	"database/sql"
	"fmt"
	"github.com/blent/beagle/pkg/tracking"
	"github.com/blent/beagle/server/storage"
	"github.com/blent/beagle/server/storage/providers/sqlite/repositories/mapping"
	"github.com/blent/beagle/server/utils"
	"github.com/pkg/errors"
	"strings"
	"sync"
)

const (
	peripheralSelectQuery       = "SELECT id, key, name, kind, enabled FROM %s"
	peripheralInsertQuery       = "INSERT INTO %s (key, name, kind, enabled) VALUES %s"
	peripheralInsertValuesQuery = "(?, ?, ?, ?)"
	peripheralUpdateQuery       = "UPDATE %s SET name=?, enabled=? WHERE id=?"
	peripheralDeleteQuery       = "DELETE FROM %s"
	peripheralCountQuery        = "SELECT COUNT(id) from %s"
)

type (
	SQLitePeripheralRepository struct {
		mu        sync.Mutex
		tableName string
		db        *sql.DB
	}
)

func NewSQLitePeripheralRepository(tableName string, db *sql.DB) *SQLitePeripheralRepository {
	return &SQLitePeripheralRepository{
		tableName: tableName,
		db:        db,
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

func (r *SQLitePeripheralRepository) Count(filter *storage.PeripheralFilter) (uint64, error) {
	queryStmt := fmt.Sprintf(peripheralCountQuery, r.tableName)
	whereKeys := make([]string, 0, 5)
	whereValues := make([]interface{}, 0, 5)

	if filter != nil {
		if filter.Status != "" {
			var enabled bool

			if filter.Status == storage.PERIPHERAL_STATUS_ENABLED {
				enabled = true
			}

			whereKeys = append(whereKeys, "enabled = ?")
			whereValues = append(whereValues, enabled)
		}

		if len(whereKeys) > 0 {
			queryStmt = fmt.Sprintf(
				"%s WHERE %s",
				queryStmt,
				strings.Join(whereKeys, " AND "),
			)
		}
	}

	stmt, err := r.db.Prepare(queryStmt)

	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	row := stmt.QueryRow(whereValues...)

	var count uint64

	err = row.Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
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

	r.mu.Lock()
	defer r.mu.Unlock()

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

	r.mu.Lock()
	defer r.mu.Unlock()

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

	var err error

	r.mu.Lock()
	defer r.mu.Unlock()

	tx, closeTx, err := storage.TryToBegin(r.db, tx)

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(
			"%s WHERE id=?",
			fmt.Sprintf(peripheralDeleteQuery, r.tableName),
		),
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

func (r *SQLitePeripheralRepository) DeleteMany(query *storage.DeletionQuery, tx *sql.Tx) error {
	if query == nil {
		return errors.New("missed query object")
	}

	if len(query.Id) == 0 {
		return errors.New("passed empty list of ids")
	}

	var err error

	r.mu.Lock()
	defer r.mu.Unlock()

	tx, closeTx, err := storage.TryToBegin(r.db, tx)

	if err != nil {
		return err
	}

	where := "WHERE id"

	if query.InRange == false {
		where += " NOT IN"
	} else {
		where += " IN"
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(
			"%s %s (%s)",
			fmt.Sprintf(peripheralDeleteQuery, r.tableName),
			where,
			utils.JoinUintSlice(query.Id, ", "),
		),
	)

	if err != nil {
		return storage.TryToRollback(tx, err, closeTx)
	}

	_, err = stmt.Exec()

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
