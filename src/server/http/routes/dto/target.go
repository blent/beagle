package dto

import (
	"github.com/blent/beagle/src/core/discovery/peripherals"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/pkg/errors"
	"strings"
)

type Target struct {
	Id          uint64        `json:"id"`
	Kind        string        `json:"kind" binding:"required"`
	Name        string        `json:"name" binding:"required"`
	Uuid        string        `json:"uuid"`
	Major       uint16        `json:"major"`
	Minor       uint16        `json:"minor"`
	Enabled     bool          `json:"enabled" binding:"required"`
	Subscribers []*Subscriber `json:"subscribers"`
}

func ToTarget(targetDto *Target) (*tracking.Target, error) {
	var err error
	var key string

	switch targetDto.Kind {
	case peripherals.PERIPHERAL_IBEACON:
		targetDto.Uuid = strings.TrimSpace(targetDto.Uuid)
		if len(targetDto.Uuid) != 32 {
			err = errors.Errorf("invalid uuid length: %d", len(targetDto.Uuid))
			break
		}

		if targetDto.Major == 0 {
			err = errors.Errorf("invalid major number: %d", targetDto.Major)
			break
		}

		if targetDto.Minor == 0 {
			err = errors.Errorf("invalid minor number: %d", targetDto.Minor)
			break
		}

		key = peripherals.CreateIBeaconUniqueKey(targetDto.Uuid, targetDto.Major, targetDto.Minor)
	default:
		err = errors.Errorf("unsupported target kind: '%s'", targetDto.Kind)
	}

	if err == nil {
		targetDto.Name = strings.TrimSpace(targetDto.Name)

		if targetDto.Name == "" {
			err = errors.New("missed name")
		}
	}

	if err != nil {
		return nil, err
	}

	var subscribers []*tracking.Subscriber

	if targetDto.Subscribers != nil && len(targetDto.Subscribers) > 0 {
		subscribers = make([]*tracking.Subscriber, 0, len(targetDto.Subscribers))

		for _, subDto := range targetDto.Subscribers {
			sub, err := ToSubscriber(subDto)

			if err != nil {
				return nil, err
			}

			subscribers = append(subscribers, sub)
		}
	}

	return &tracking.Target{
		Id:          targetDto.Id,
		Key:         key,
		Name:        targetDto.Name,
		Kind:        targetDto.Kind,
		Enabled:     targetDto.Enabled,
		Subscribers: subscribers,
	}, nil
}

func FromTarget(target *tracking.Target) (*Target, error) {
	var targetDto *Target
	var err error

	switch target.Kind {
	case peripherals.PERIPHERAL_IBEACON:
		uuid, major, minor, err := peripherals.ParseIBeaconUniqueKey(target.Key)

		if err == nil {
			targetDto = &Target{
				Uuid:  uuid,
				Major: major,
				Minor: minor,
			}
		}
	default:
		err = errors.Errorf("unsupported target kind: '%s'", targetDto.Kind)
	}

	if err != nil {
		return nil, err
	}

	targetDto.Id = target.Id
	targetDto.Kind = target.Kind
	targetDto.Name = target.Name
	targetDto.Enabled = target.Enabled

	var subscribers []*Subscriber

	if target.Subscribers != nil && len(target.Subscribers) > 0 {
		subscribers = make([]*Subscriber, 0, len(targetDto.Subscribers))

		for _, subscriber := range target.Subscribers {
			subscribers = append(subscribers, &Subscriber{
				Subscriber: subscriber,
			})
		}
	}

	targetDto.Subscribers = subscribers

	return targetDto, nil
}
