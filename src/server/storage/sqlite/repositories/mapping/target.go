package mapping

import (
	"database/sql"
	"github.com/blent/beagle/src/core/tracking"
)

type (
	DataRows interface {
		DataRow
		Next() bool
		Close() error
	}

	DataRow interface {
		Scan(...interface{}) error
	}
)

func ToTarget(row DataRow) (*tracking.Target, error) {
	var id uint64
	var key string
	var name string
	var kind string
	var enabled int

	if err := row.Scan(&id, &key, &name, &kind, &enabled); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &tracking.Target{
		Id:      id,
		Key:     key,
		Name:    name,
		Kind:    kind,
		Enabled: enabled == 1,
	}, nil
}

func ToTargets(rows DataRows, size uint64) ([]*tracking.Target, error) {
	results := make([]*tracking.Target, 0, size)
	var err error
	defer rows.Close()

	for rows.Next() {
		target, parseErr := ToTarget(rows)

		if parseErr != nil {
			err = parseErr
			break
		}

		results = append(results, target)
	}

	if err != nil {
		return nil, err
	}

	return results, nil
}
