package tracking

import "github.com/jinzhu/gorm"

type Target struct {
	gorm.Model
	Key         string        `json:"key" gorm:"unique_index"`
	Name        string        `json:"name"`
	Kind        string        `json:"kind" gorm:"index"`
	Subscribers []*Subscriber `json:"subscribers" gorm:"many2many:target_subscribers;"`
	Enabled     bool          `json:"enabled"`
}
