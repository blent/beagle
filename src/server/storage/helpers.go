package storage

import (
	"database/sql"
	"fmt"
)

func TryToBegin(db *sql.DB, tx *sql.Tx) (*sql.Tx, bool, error) {
	if tx != nil {
		return tx, false, nil
	}

	tx, err := db.Begin()

	if err != nil {
		return nil, false, err
	}

	return tx, true, nil
}

func TryToCommit(tx *sql.Tx, close bool) error {
	if close {
		return tx.Commit()
	}

	return nil
}

func TryToRollback(tx *sql.Tx, reason error, close bool) error {
	if close {
		rollbackErr := tx.Rollback()

		if rollbackErr != nil {
			return fmt.Errorf("%s:%s", rollbackErr.Error(), reason.Error())
		}
	}

	return reason
}
