package activity

import (
	"time"
)

type Record struct {
	Key        string    `json:"key"`
	Kind       string    `json:"kind"`
	Proximity  string    `json:"proximity"`
	Registered bool      `json:"registered"`
	Time       time.Time `json:"time"`
}
