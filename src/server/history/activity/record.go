package activity

import (
	"time"
)

type Record struct {
	Id        int
	Time      time.Time `json:"time"`
	Key       string    `json:"key"`
	Kind      string    `json:"kind"`
	Proximity string    `json:"proximity"`
}
