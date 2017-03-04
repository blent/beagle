package mapping

import (
	"database/sql"
	"github.com/blent/beagle/src/core/notification"
	"github.com/blent/beagle/src/server/storage"
)

func ToSubscriber(row DataRow) (*notification.Subscriber, error) {
	var id uint64
	var name string
	var event string
	var enabled uint64

	var endpointId uint64
	var endpointName string
	var endpointUrl string
	var endpointMethod string
	var endpointHeaders notification.Headers

	if err := row.Scan(
		&id,
		&name,
		&event,
		&enabled,
		&endpointId,
		&endpointName,
		&endpointUrl,
		&endpointMethod,
		&endpointHeaders,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &notification.Subscriber{
		Id:      id,
		Name:    name,
		Event:   event,
		Enabled: enabled > 0,
		Endpoint: &notification.Endpoint{
			Id:      endpointId,
			Name:    endpointName,
			Url:     endpointUrl,
			Method:  endpointMethod,
			Headers: endpointHeaders,
		},
	}, nil
}

func ToSubscribers(rows DataRows, query *storage.SubscriberQuery) ([]*notification.Subscriber, error) {
	results := make([]*notification.Subscriber, 0, query.Take)
	var err error
	defer rows.Close()

	for rows.Next() {
		target, parseErr := ToSubscriber(rows)

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
