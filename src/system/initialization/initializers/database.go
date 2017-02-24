package initializers

import (
	"github.com/jinzhu/gorm"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/system/history/activity"
)

type DatabaseInitializer struct {
	logger *logging.Logger
	db     *gorm.DB
}

func NewDatabaseInitializer(logger *logging.Logger, db *gorm.DB) *DatabaseInitializer {
	return &DatabaseInitializer{logger, db}
}

func (init *DatabaseInitializer) Run() error {
	init.db.LogMode(true)

	res := init.db.AutoMigrate(
		&tracking.Target{},
		&tracking.Subscriber{},
		&activity.Record{},
	)

	if res.Error != nil {
		return res.Error
	}

	return nil
}
