package delivery

import (
	"github.com/blent/beagle/src/core/discovery/peripherals"
	"github.com/blent/beagle/src/core/tracking"
)

type (
	Event struct {
		name        string
		targetName  string
		peripheral  peripherals.Peripheral
		subscribers []*tracking.Subscriber
	}
)

func NewEvent(name, targetName string, peripheral peripherals.Peripheral, subscribers []*tracking.Subscriber) *Event {
	return &Event{
		name,
		targetName,
		peripheral,
		subscribers,
	}
}

func (event *Event) Name() string {
	return event.name
}

func (event *Event) TargetName() string {
	return event.targetName
}

func (event *Event) Peripheral() peripherals.Peripheral {
	return event.peripheral
}

func (event *Event) Subscribers() []*tracking.Subscriber {
	return event.subscribers
}
