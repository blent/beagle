package delivery

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Record struct {
	gorm.Model
	Time       time.Time `json:"time"`
	Key        string    `json:"key" gorm:"index"`
	Kind       string    `json:"kind" gorm:"index"`
	Subscriber string    `json:"subscriber" gorm:"index"`
	Delivered  bool      `json:"delivered"`
}

func (Record) TableName() string {
	return "delivery_history"
}
