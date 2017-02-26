package delivery

import (
	"time"
)

type Record struct {
	Id         int
	Time       time.Time `json:"time"`
	Key        string    `json:"key"`
	Kind       string    `json:"kind"`
	Subscriber string    `json:"subscriber"`
	Delivered  bool      `json:"delivered"`
}
