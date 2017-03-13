package sqlite

import (
	"database/sql"
	"fmt"
)

type tableCreator func(tx *sql.Tx) error

func initialize(tx *sql.Tx) (bool, error) {
	tables, err := getTableCreators(tx)

	if err != nil {
		return false, err
	}

	if len(tables) == 0 {
		return false, nil
	}

	for _, table := range tables {
		if err = table(tx); err != nil {
			break
		}
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func getTableCreators(tx *sql.Tx) (map[string]tableCreator, error) {
	rows, err := tx.Query("SELECT name FROM sqlite_master WHERE type='table'")

	if err != nil {
		return nil, err
	}

	tables := make(map[string]tableCreator)
	tables[peripheralTableName] = createPeripheralsTable
	tables[endpointTableName] = createEndpointsTable
	tables[subscriberTableName] = createSubscribersTable

	for rows.Next() {
		var name string

		err = rows.Scan(&name)

		if err != nil {
			break
		}

		delete(tables, name)
	}

	if err != nil {
		return nil, err
	}

	return tables, nil
}

func execQueries(tx *sql.Tx, queries []string) error {
	var err error

	for _, query := range queries {
		if _, err = tx.Exec(query); err != nil {
			break
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func createPeripheralsTable(tx *sql.Tx) error {
	return execQueries(tx, []string{
		fmt.Sprintf(
			"CREATE TABLE %s("+
				"id INTEGER NOT NULL PRIMARY KEY,"+
				"key TEXT NOT NULL,"+
				"name TEXT NOT NULL,"+
				"kind TEXT NOT NULL,"+
				"enabled INTEGER NOT NULL"+
				");",
			peripheralTableName,
		),
		fmt.Sprintf(
			"CREATE UNIQUE INDEX %s_key_idx on %s(key);",
			peripheralTableName,
			peripheralTableName,
		),
		fmt.Sprintf(
			"CREATE UNIQUE INDEX %s_name_idx on %s(name);",
			peripheralTableName,
			peripheralTableName,
		),
	})
}

func createEndpointsTable(tx *sql.Tx) error {
	return execQueries(tx, []string{
		fmt.Sprintf(
			"CREATE TABLE %s("+
				"id INTEGER NOT NULL PRIMARY KEY,"+
				"name TEXT NOT NULL,"+
				"url TEXT NOT NULL,"+
				"method TEXT NOT NULL,"+
				"headers TEXT"+
				");",
			endpointTableName,
		),
		fmt.Sprintf(
			"CREATE UNIQUE INDEX %s_name_idx on %s(name);",
			endpointTableName,
			endpointTableName,
		),
	})
}

func createSubscribersTable(tx *sql.Tx) error {
	return execQueries(tx, []string{
		fmt.Sprintf(
			"CREATE TABLE %s("+
				"id INTEGER NOT NULL PRIMARY KEY,"+
				"name TEXT NOT NULL,"+
				"event TEXT NOT NULL,"+
				"enabled INTEGER NOT NULL,"+
				"target_id INTEGER REFERENCES %s(id) ON DELETE CASCADE,"+
				"endpoint_id INTEGER REFERENCES %s(id) ON DELETE CASCADE"+
				");",
			subscriberTableName,
			peripheralTableName,
			endpointTableName,
		),
		fmt.Sprintf(
			"CREATE INDEX %s_target_idx on %s(target_id);",
			subscriberTableName,
			subscriberTableName,
		),
		fmt.Sprintf(
			"CREATE INDEX %s_endpoint_idx on %s(endpoint_id);",
			subscriberTableName,
			subscriberTableName,
		),
	})
}
