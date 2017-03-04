package dto

import (
	"github.com/blent/beagle/src/core/notification"
	"github.com/go-errors/errors"
	"strings"
)

type Subscriber struct {
	*notification.Subscriber
}

func ToSubscriber(subDto *Subscriber) (*notification.Subscriber, error) {
	subDto.Name = strings.TrimSpace(subDto.Name)

	if subDto.Name == "" {
		return nil, errors.New("missed subscriber name")
	}

	subDto.Event = strings.TrimSpace(subDto.Event)

	if subDto.Event == "" {
		return nil, errors.New("missed subscriber event")
	}

	return &notification.Subscriber{
		Id:       subDto.Id,
		Name:     subDto.Name,
		Event:    subDto.Event,
		Enabled:  subDto.Enabled,
		Endpoint: subDto.Endpoint,
	}, nil
}

func FromSubscriber(sub *notification.Subscriber) (*Subscriber, error) {
	return &Subscriber{
		Subscriber: sub,
	}, nil
}
