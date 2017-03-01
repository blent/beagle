package dto

import (
	"github.com/blent/beagle/src/core/tracking"
	"github.com/go-errors/errors"
	"strings"
)

type Subscriber struct {
	*tracking.Subscriber
}

func ToSubscriber(subDto *Subscriber) (*tracking.Subscriber, error) {
	subDto.Name = strings.TrimSpace(subDto.Name)

	if subDto.Name == "" {
		return nil, errors.New("missed subscriber name")
	}

	subDto.Event = strings.TrimSpace(subDto.Event)

	if subDto.Event == "" {
		return nil, errors.New("missed subscriber event")
	}

	subDto.Method = strings.TrimSpace(subDto.Method)

	if subDto.Method == "" {
		return nil, errors.New("missed subscriber method")
	}

	subDto.Url = strings.TrimSpace(subDto.Url)

	if subDto.Url == "" {
		return nil, errors.New("missed subscriber url")
	}

	return &tracking.Subscriber{
		Id:      subDto.Id,
		Name:    subDto.Name,
		Event:   subDto.Event,
		Method:  subDto.Method,
		Url:     subDto.Url,
		Enabled: subDto.Enabled,
		Headers: subDto.Headers,
		Data:    subDto.Data,
	}, nil
}
