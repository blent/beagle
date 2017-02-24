package tracking

import "time"

type Settings struct {
	Ttl       time.Duration
	Heartbeat time.Duration
}
