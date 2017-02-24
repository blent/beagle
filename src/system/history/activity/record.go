package activity

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Record struct {
	gorm.Model
	Time      time.Time `json:"time"`
	Key       string    `json:"key" gorm:"index"`
	Kind      string    `json:"kind" gorm:"index"`
	Proximity string    `json:"proximity" gorm:"-"`
}

func (Record) TableName() string {
	return "activity_history"
}
