package server

import (
	"github.com/blent/beagle/pkg/notification"
	"github.com/blent/beagle/pkg/tracking"
	"github.com/blent/beagle/server/storage"
)

type Registry struct {
	db *storage.Manager
}

func NewRegistry(db *storage.Manager) (*Registry, error) {
	return &Registry{db}, nil
}

func (r *Registry) FindTarget(key string) (*tracking.Peripheral, error) {
	return r.db.GetPeripheralByKey(key)
}

func (r *Registry) FindSubscribers(targetId uint64, events ...string) ([]*notification.Subscriber, error) {
	return r.db.GetPeripheralSubscribersByEvent(
		targetId,
		events,
		storage.PERIPHERAL_STATUS_ENABLED,
	)
}
