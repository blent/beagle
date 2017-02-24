package storage

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/blent/beagle/src/core/tracking"
)

const (
	STATUS_ALL = iota
	STATUS_ENABLED
	STATUS_DISABLED
)

type (
	TargetQuery struct {
		Take   int64
		Skip   int64
		Status uint64
	}
	TargetRepository interface {
		GetById(uint64) (*tracking.Target, error)
		GetByKey(string) (*tracking.Target, error)
		Query(*TargetQuery) ([]*tracking.Target, error)
		Create(*tracking.Target) error
		Update(*tracking.Target) error
		Delete(uint64) error
		DeleteMany([]uint64) error
	}

	SubscriberRepository interface {
		GetByName(string) (*tracking.Subscriber, error)
	}
)

func NewDatabase(settings *Settings) (*gorm.DB, error) {
	return gorm.Open(settings.Dialect, settings.ConnectionString)
}
