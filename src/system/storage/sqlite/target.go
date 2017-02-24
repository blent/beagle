package sqlite

import (
	"github.com/jinzhu/gorm"
	"github.com/blent/beagle/src/core/tracking"
)

type (
	TargetRepository struct {
		db *gorm.DB
	}
)

func NewSQLiteTargetRepository(db *gorm.DB) *TargetRepository {
	return &TargetRepository{db}
}

func (repo *TargetRepository) FindByKey(key string) (*tracking.Target, error) {
	var target tracking.Target

	resp := repo.db.Where("key = ?", key).First(&target)

	if resp.Error != nil {
		return nil, resp.Error
	}

	if resp.RecordNotFound() {
		return nil, nil
	}

	return &target, nil
}
