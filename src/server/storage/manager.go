package storage

import (
	"database/sql"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/notification"
	"github.com/blent/beagle/src/core/tracking"
)

type Manager struct {
	logger      *logging.Logger
	db          *sql.DB
	targets     TargetRepository
	subscribers SubscriberRepository
	endpoints   EndpointRepository
}

func NewManager(logger *logging.Logger, provider Provider) *Manager {
	return &Manager{
		logger:      logger,
		db:          provider.GetConnection(),
		targets:     provider.GetTargetRepository(),
		subscribers: provider.GetSubscriberRepository(),
		endpoints:   provider.GetEndpointRepository(),
	}
}

func (m *Manager) FindTargets(query *TargetQuery) ([]*tracking.Target, error) {
	return m.targets.Find(query)
}

func (m *Manager) GetTarget(id uint64) (*tracking.Target, error) {
	return m.targets.Get(id)
}

func (m *Manager) GetTargetByKey(key string) (*tracking.Target, error) {
	return m.targets.GetByKey(key)
}

func (m *Manager) GetTargetWithSubscribers(id uint64) (*tracking.Target, []*notification.Subscriber, error) {
	target, err := m.targets.Get(id)

	if err != nil {
		return nil, nil, err
	}

	if target == nil {
		return nil, nil, nil
	}

	subscribers, err := m.subscribers.Find(NewSubscriberQuery(0, 0, id, "*"))

	if err != nil {
		return nil, nil, err
	}

	return target, subscribers, nil
}

func (m *Manager) GetTargetSubscribersByEvent(targetId uint64, eventName string) ([]*notification.Subscriber, error) {
	return m.subscribers.Find(NewSubscriberQuery(0, 0, targetId, eventName))
}

func (m *Manager) CreateTarget(target *tracking.Target, subscribers []*notification.Subscriber) (uint64, error) {
	tx, err := m.db.Begin()

	if err != nil {
		return 0, err
	}

	id, err := m.targets.Create(target, tx)

	if err != nil {
		return 0, TryToRollback(tx, err, true)
	}

	err = m.subscribers.CreateMany(subscribers, id, tx)

	if err != nil {
		return 0, TryToRollback(tx, err, true)
	}

	err = tx.Commit()

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *Manager) UpdateTarget(target *tracking.Target, subscribers []*notification.Subscriber) error {
	tx, err := m.db.Begin()

	if err != nil {
		return err
	}

	err = m.targets.Update(target, tx)

	if err != nil {
		return TryToRollback(tx, err, true)
	}

	if subscribers != nil && len(subscribers) > 0 {
		update := make([]*notification.Subscriber, 0, len(subscribers))
		create := make([]*notification.Subscriber, 0, len(subscribers))

		for _, subscriber := range subscribers {
			if subscriber.Id == 0 {
				create = append(create, subscriber)
			} else {
				update = append(update, subscriber)
			}
		}

		err = m.subscribers.UpdateMany(update, tx)

		if err != nil {
			return TryToRollback(tx, err, true)
		}

		err = m.subscribers.CreateMany(create, target.Id, tx)

		if err != nil {
			return TryToRollback(tx, err, true)
		}
	}

	return TryToCommit(tx, true)
}

func (m *Manager) DeleteTarget(id uint64) error {
	return m.targets.Delete(id, nil)
}

func (m *Manager) FindEndpoints(query *EndpointQuery) ([]*notification.Endpoint, error) {
	return m.endpoints.Find(query)
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
