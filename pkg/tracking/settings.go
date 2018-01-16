package tracking

import "time"

type Settings struct {
	Ttl       time.Duration
	Heartbeat time.Duration
}

func (s *Settings) Equals(other *Settings) bool {
	if other == nil {
		return false
	}

	if s.Ttl != other.Ttl {
		return false
	}

	if s.Heartbeat != other.Heartbeat {
		return false
	}

	return true
}
