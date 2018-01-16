package tracking

import (
	"github.com/blent/beagle/pkg/discovery/peripherals"
	"time"
)

type (
	Track struct {
		peripheral peripherals.Peripheral
		ttl        time.Duration
		lastSeen   time.Time
	}
)

func NewTrack(peripheral peripherals.Peripheral, ttl time.Duration) *Track {
	return &Track{
		peripheral: peripheral,
		ttl:        ttl,
		lastSeen:   time.Now(),
	}
}

func (record *Track) Peripheral() peripherals.Peripheral {
	return record.peripheral
}

func (record *Track) Update() {
	record.lastSeen = time.Now()
}

func (record *Track) IsActive() bool {
	return record.ttl > time.Since(record.lastSeen)
}
