package notification

import (
	"github.com/blent/beagle/pkg/discovery/peripherals"
)

type (
	Message struct {
		eventName   string
		targetName  string
		peripheral  peripherals.Peripheral
		subscribers []*Subscriber
	}
)

func NewMessage(eventName, targetName string, peripheral peripherals.Peripheral, subscribers []*Subscriber) *Message {
	return &Message{
		eventName,
		targetName,
		peripheral,
		subscribers,
	}
}

func (event *Message) EventName() string {
	return event.eventName
}

func (event *Message) TargetName() string {
	return event.targetName
}

func (event *Message) Peripheral() peripherals.Peripheral {
	return event.peripheral
}

func (event *Message) Subscribers() []*Subscriber {
	return event.subscribers
}
