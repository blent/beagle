package repositories

import (
	"database/sql"
	"fmt"
	"github.com/blent/beagle/src/core/notification"
	"github.com/blent/beagle/src/server/storage"
	"github.com/blent/beagle/src/server/storage/providers/sqlite/repositories/mapping"
	"github.com/blent/beagle/src/server/utils"
	"github.com/pkg/errors"
	"strings"
	"sync"
)

const (
	subscriberSelectQuery = "SELECT " +
		"t1.id as t1_id, " +
		"t1.name as t1_name, " +
		"t1.event as t1_event, " +
		"t1.enabled as t1_enabled, " +
		"t2.id AS t2_id, " +
		"t2.name AS t2_name, " +
		"t2.url AS t2_url, " +
		"t2.method AS t2_method, " +
		"t2.headers AS t2_headers " +
		"FROM %s AS t1 " +
		"INNER JOIN %s AS t2 ON t1.endpoint_id = t2.id "
	subscriberInsertQuery       = "INSERT INTO %s (name, event, enabled, endpoint_id, target_id) VALUES %s"
	subscriberInsertValuesQuery = "(?, ?, ?, ?, ?)"
	subscriberUpdateQuery       = "UPDATE %s SET name=?, event=?, enabled=? WHERE id=?"
	subscriberDeleteQuery       = "DELETE FROM %s"
	subscriberCountQuery        = "SELECT COUNT(id) FROM %s"
)

type SQLiteSubscriberRepository struct {
	mu                sync.Mutex
	tableName         string
	endpointTableName string
	db                *sql.DB
}

func NewSQLiteSubscriberRepository(tableName, endpointTableName string, db *sql.DB) *SQLiteSubscriberRepository {
	return &SQLiteSubscriberRepository{
		tableName:         tableName,
		endpointTableName: endpointTableName,
		db:                db,
	}
}

func (r *SQLiteSubscriberRepository) Get(id uint64) (*notification.Subscriber, error) {
	if id == 0 {
		return nil, errors.New("id must be greater than 0")
	}

	stmt, err := r.db.Prepare(
		fmt.Sprintf(
			"%s WHERE t1.id=? LIMIT 1",
			fmt.Sprintf(
				subscriberSelectQuery,
				r.tableName,
				r.endpointTableName,
			),
		),
	)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return mapping.ToSubscriber(stmt.QueryRow(id))
}

func (r *SQLiteSubscriberRepository) Count(filter *storage.SubscriberFilter) (uint64, error) {
	queryStmt := fmt.Sprintf(subscriberCountQuery, r.tableName)
	whereKeys := make([]string, 0, 5)
	whereValues := make([]interface{}, 0, 5)

	if filter != nil {
		if filter.Event != "" {
			whereKeys = append(whereKeys, "event = ?")
			whereValues = append(whereValues, filter.Event)
		}

		if filter.TargetId > 0 {
			whereKeys = append(whereKeys, "target_id = ?")
			whereValues = append(whereValues, filter.TargetId)
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

func (r *SQLiteSubscriberRepository) Find(query *storage.SubscriberQuery) ([]*notification.Subscriber, error) {
	if query == nil {
		return nil, errors.New("query object is missed")
	}

	var queryStmt string

	selectQuery := subscriberSelectQuery
	args := make([]interface{}, 0, 3)
	where := make([]string, 0, 10)

	if query.TargetId > 0 {
		args = append(args, query.TargetId)
		where = append(where, "t1.target_id = ?")
	}

	if query.Event != "" && query.Event != "*" {
		where = append(where, "t1.event = ?")
		args = append(args, query.Event)
	}

	if len(where) > 0 {
		selectQuery += "WHERE " + strings.Join(where, " AND ")
	}

	selectQuery += " ORDER BY t1.id"

	if query != nil && query.Take > 0 {
		args = append(args, query.Take, query.Skip)

		queryStmt = fmt.Sprintf(
			"%s LIMIT ? OFFSET ?",
			fmt.Sprintf(
				selectQuery,
				r.tableName,
				r.endpointTableName,
			),
		)
	} else {
		queryStmt = fmt.Sprintf(selectQuery, r.tableName, r.endpointTableName)
	}

	stmt, err := r.db.Prepare(queryStmt)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(args...)

	if err != nil {
		return nil, err
	}

	return mapping.ToSubscribers(rows, query)
}

func (r *SQLiteSubscriberRepository) Create(subscriber *notification.Subscriber, targetId uint64, tx *sql.Tx) (uint64, error) {
	var id int64
	var err error

	if err := r.validate(subscriber, true); err != nil {
		return 0, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	tx, closeTx, err := storage.TryToBegin(r.db, tx)

	if err != nil {
		return 0, err
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(subscriberInsertQuery, r.tableName, subscriberInsertValuesQuery),
	)

	if err != nil {
		return 0, storage.TryToRollback(tx, err, closeTx)
	}

	res, err := stmt.Exec(
		subscriber.Name,
		subscriber.Event,
		boolToInt(subscriber.Enabled),
		subscriber.Endpoint.Id,
		targetId,
	)

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

	return uint64(id), err
}

func (r *SQLiteSubscriberRepository) CreateMany(subscribers []*notification.Subscriber, targetId uint64, tx *sql.Tx) error {
	if subscribers == nil {
		return errors.New("subscribers missed")
	}

	var err error
	valueStrings := make([]string, 0, len(subscribers))
	valueArgs := make([]interface{}, 0, len(subscribers)*5)

	for _, subscriber := range subscribers {
		err := r.validate(subscriber, true)

		if err != nil {
			break
		}

		// name, event, enabled, endpoint_id, target_id
		valueStrings = append(valueStrings, subscriberInsertValuesQuery)
		valueArgs = append(
			valueArgs,
			subscriber.Name,
			subscriber.Event,
			boolToInt(subscriber.Enabled),
			subscriber.Endpoint.Id,
			targetId,
		)
	}

	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	tx, closeTx, err := storage.TryToBegin(r.db, tx)

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(
			subscriberInsertQuery,
			r.tableName,
			strings.Join(valueStrings, ","),
		),
	)

	if err != nil {
		return storage.TryToRollback(tx, err, closeTx)
	}

	_, err = stmt.Exec(valueArgs...)

	if err != nil {
		return storage.TryToRollback(tx, err, closeTx)
	}

	return storage.TryToCommit(tx, closeTx)
}

func (r *SQLiteSubscriberRepository) Update(subscriber *notification.Subscriber, tx *sql.Tx) error {
	if err := r.validate(subscriber, false); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	tx, closeTx, err := storage.TryToBegin(r.db, tx)

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(subscriberUpdateQuery, r.tableName),
	)

	if err != nil {
		return storage.TryToRollback(tx, err, closeTx)
	}

	err = r.doUpdate(stmt, subscriber)

	if err != nil {
		return storage.TryToRollback(tx, err, closeTx)
	}

	return storage.TryToCommit(tx, closeTx)
}

func (r *SQLiteSubscriberRepository) UpdateMany(subscribers []*notification.Subscriber, tx *sql.Tx) error {
	if subscribers == nil {
		return errors.New("missed subscribers")
	}

	var err error

	for _, subscriber := range subscribers {
		err = r.validate(subscriber, false)

		if err != nil {
			break
		}
	}

	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	tx, closeTx, err := storage.TryToBegin(r.db, tx)

	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(
		fmt.Sprintf(subscriberUpdateQuery, r.tableName),
	)

	if err != nil {
		return storage.TryToRollback(tx, err, closeTx)
	}

	for _, subscriber := range subscribers {
		err = r.doUpdate(stmt, subscriber)

		if err != nil {
			break
		}
	}

	if err != nil {
		return storage.TryToRollback(tx, err, closeTx)
	}

	return storage.TryToCommit(tx, closeTx)
}

func (r *SQLiteSubscriberRepository) Delete(id uint64, tx *sql.Tx) error {
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
			fmt.Sprintf(subscriberDeleteQuery, r.tableName),
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

func (r *SQLiteSubscriberRepository) DeleteMany(query *storage.DeletionQuery, tx *sql.Tx) error {
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
			fmt.Sprintf(subscriberDeleteQuery, r.tableName),
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

func (r *SQLiteSubscriberRepository) doUpdate(stmt *sql.Stmt, subscriber *notification.Subscriber) error {
	_, err := stmt.Exec(subscriber.Name, subscriber.Event, boolToInt(subscriber.Enabled), subscriber.Id)

	return err
}

func (r *SQLiteSubscriberRepository) validate(subscriber *notification.Subscriber, isNew bool) error {
	if subscriber == nil {
		return errors.New("subscriber missed")
	}

	if isNew {
		if subscriber.Id > 0 {
			return errors.New("subscriber already created")
		}
	} else {
		if subscriber.Id == 0 {
			return errors.New("subscriber already created")
		}
	}

	if subscriber.Endpoint == nil {
		return errors.New("subscriber must contain an endpoint")
	}

	if subscriber.Endpoint.Id == 0 {
		return errors.New("subscriber must contain existing endpoint")
	}

	return nil
}
