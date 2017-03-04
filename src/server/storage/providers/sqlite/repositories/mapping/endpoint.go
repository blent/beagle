package mapping

import (
	"database/sql"
	"github.com/blent/beagle/src/core/notification"
)

func ToEndpoint(row DataRow) (*notification.Endpoint, error) {
	var id uint64
	var name string
	var url string
	var method string
	headers := notification.Headers{}

	if err := row.Scan(&id, &name, &url, &method, &headers); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &notification.Endpoint{
		Id:      id,
		Name:    name,
		Url:     url,
		Method:  method,
		Headers: headers,
	}, nil
}

func ToEndpoints(rows DataRows, size uint64) ([]*notification.Endpoint, error) {
	results := make([]*notification.Endpoint, 0, size)
	var err error
	defer rows.Close()

	for rows.Next() {
		target, parseErr := ToEndpoint(rows)

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
