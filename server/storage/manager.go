package storage

import (
	"database/sql"
	"github.com/blent/beagle/pkg/notification"
	"github.com/blent/beagle/pkg/tracking"
	"go.uber.org/zap"
)

type Manager struct {
	logger      *zap.Logger
	db          *sql.DB
	peripherals PeripheralRepository
	subscribers SubscriberRepository
	endpoints   EndpointRepository
}

func NewManager(logger *zap.Logger, provider Provider) *Manager {
	return &Manager{
		logger:      logger,
		db:          provider.GetConnection(),
		peripherals: provider.GetPeripheralRepository(),
		subscribers: provider.GetSubscriberRepository(),
		endpoints:   provider.GetEndpointRepository(),
	}
}

func (m *Manager) FindPeripherals(query *PeripheralQuery) ([]*tracking.Peripheral, uint64, error) {
	res, err := m.peripherals.Find(query)

	if err != nil {
		return nil, 0, err
	}

	count, err := m.peripherals.Count(query.PeripheralFilter)

	if err != nil {
		return nil, 0, err
	}

	return res, count, nil
}

func (m *Manager) GetPeripheral(id uint64) (*tracking.Peripheral, error) {
	return m.peripherals.Get(id)
}

func (m *Manager) GetPeripheralByKey(key string) (*tracking.Peripheral, error) {
	return m.peripherals.GetByKey(key)
}

func (m *Manager) GetPeripheralWithSubscribers(id uint64) (*tracking.Peripheral, []*notification.Subscriber, error) {
	target, err := m.peripherals.Get(id)

	if err != nil {
		return nil, nil, err
	}

	if target == nil {
		return nil, nil, nil
	}

	subscribers, err := m.subscribers.Find(NewSubscriberQuery(
		0,
		0,
		id,
		nil,
		PERIPHERAL_STATUS_ANY,
	))

	if err != nil {
		return nil, nil, err
	}

	return target, subscribers, nil
}

func (m *Manager) GetPeripheralSubscribersByEvent(targetId uint64, eventNames []string, status string) ([]*notification.Subscriber, error) {
	return m.subscribers.Find(NewSubscriberQuery(
		0,
		0,
		targetId,
		eventNames,
		status,
	))
}

func (m *Manager) CreatePeripheral(target *tracking.Peripheral, subscribers []*notification.Subscriber) (uint64, error) {
	tx, err := m.db.Begin()

	if err != nil {
		return 0, err
	}

	id, err := m.peripherals.Create(target, tx)

	if err != nil {
		return 0, TryToRollback(tx, err, true)
	}

	if subscribers != nil && len(subscribers) > 0 {
		err = m.subscribers.CreateMany(subscribers, id, tx)

		if err != nil {
			return 0, TryToRollback(tx, err, true)
		}
	}

	err = tx.Commit()

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *Manager) UpdatePeripheral(target *tracking.Peripheral, subscribers []*notification.Subscriber) error {
	tx, err := m.db.Begin()

	if err != nil {
		return err
	}

	err = m.peripherals.Update(target, tx)

	if err != nil {
		return TryToRollback(tx, err, true)
	}

	if subscribers != nil && len(subscribers) > 0 {
		update := make([]*notification.Subscriber, 0, len(subscribers))
		create := make([]*notification.Subscriber, 0, len(subscribers))
		existingIds := make([]uint64, 0, len(subscribers))

		for _, subscriber := range subscribers {
			if subscriber.Id == 0 {
				create = append(create, subscriber)
			} else {
				update = append(update, subscriber)
				existingIds = append(existingIds, subscriber.Id)
			}
		}

		if len(update) > 0 {
			err = m.subscribers.UpdateMany(update, tx)

			if err != nil {
				return TryToRollback(tx, err, true)
			}

			// delete those that are not part of the payload
			err = m.subscribers.DeleteMany(&DeletionQuery{
				Id:      existingIds,
				InRange: false,
			}, tx)

			if err != nil {
				return TryToRollback(tx, err, true)
			}
		}

		if len(create) > 0 {
			err = m.subscribers.CreateMany(create, target.Id, tx)

			if err != nil {
				return TryToRollback(tx, err, true)
			}
		}
	}

	return TryToCommit(tx, true)
}

func (m *Manager) DeletePeripheral(id uint64) error {
	return m.peripherals.Delete(id, nil)
}

func (m *Manager) DeletePeripherals(ids []uint64) error {
	return m.peripherals.DeleteMany(&DeletionQuery{
		Id:      ids,
		InRange: true,
	}, nil)
}

func (m *Manager) FindEndpoints(query *EndpointQuery) ([]*notification.Endpoint, uint64, error) {
	res, err := m.endpoints.Find(query)

	if err != nil {
		return nil, 0, err
	}

	count, err := m.endpoints.Count(query.EndpointFilter)

	if err != nil {
		return nil, 0, err
	}

	return res, count, nil
}

func (m *Manager) GetEndpoint(id uint64) (*notification.Endpoint, error) {
	return m.endpoints.Get(id)
}

func (m *Manager) CreateEndpoint(endpoint *notification.Endpoint) (uint64, error) {
	return m.endpoints.Create(endpoint, nil)
}

func (m *Manager) UpdateEndpoint(endpoint *notification.Endpoint) error {
	return m.endpoints.Update(endpoint, nil)
}

func (m *Manager) DeleteEndpoint(id uint64) error {
	return m.endpoints.Delete(id, nil)
}

func (m *Manager) DeleteEndpoints(ids []uint64) error {
	return m.endpoints.DeleteMany(&DeletionQuery{
		Id:      ids,
		InRange: true,
	}, nil)
}
