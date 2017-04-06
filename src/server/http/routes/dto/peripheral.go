package dto

import (
	"github.com/blent/beagle/src/core/discovery/peripherals"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/server/utils"
	"github.com/pkg/errors"
	"strings"
)

type (
	Peripheral interface {
		GetId() uint64
		GetKind() string
		GetName() string
		GetEnabled() bool
		GetSubscribers() []*Subscriber
		SetSubscribers([]*Subscriber)
	}

	GenericPeripheral struct {
		Id          uint64        `json:"id"`
		Kind        string        `json:"kind" binding:"required"`
		Name        string        `json:"name" binding:"required"`
		Enabled     bool          `json:"enabled" binding:"required"`
		Subscribers []*Subscriber `json:"subscribers"`
	}
	IBeaconPeripheral struct {
		*GenericPeripheral
		Uuid  string `json:"uuid"`
		Major uint16 `json:"major"`
		Minor uint16 `json:"minor"`
	}
)

func (p *GenericPeripheral) GetId() uint64 {
	return p.Id
}

func (p *GenericPeripheral) GetKind() string {
	return p.Kind
}

func (p *GenericPeripheral) GetName() string {
	return p.Name
}

func (p *GenericPeripheral) GetEnabled() bool {
	return p.Enabled
}

func (p *GenericPeripheral) GetSubscribers() []*Subscriber {
	return p.Subscribers
}

func (p *GenericPeripheral) SetSubscribers(subs []*Subscriber) {
	p.Subscribers = subs
}

func ToPeripheral(input Peripheral) (*tracking.Peripheral, error) {
	var err error
	var key string

	switch input.GetKind() {
	case peripherals.PERIPHERAL_IBEACON:
		targetDto, ok := input.(*IBeaconPeripheral)

		if !ok {
			err = errors.New("invalid dto type")
			break
		}

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
		err = errors.Errorf("unsupported peripheral kind: '%s'", input.GetKind())
	}

	if err != nil {
		return nil, err
	}

	return &tracking.Peripheral{
		Id:      input.GetId(),
		Key:     key,
		Name:    input.GetName(),
		Kind:    input.GetKind(),
		Enabled: input.GetEnabled(),
	}, nil
}

func FromPeripheralMap(input map[string]interface{}) (Peripheral, error) {
	if input == nil {
		return nil, errors.New("missed input")
	}

	if len(input) == 0 {
		return nil, errors.New("invalid peripheral")
	}

	kindI := input["kind"]
	var kind string

	kind, ok := kindI.(string)

	if !ok {
		return nil, errors.New("failed to parse peripheral kind")
	}

	var err error

	switch kind {
	case peripherals.PERIPHERAL_IBEACON:
		ibeacon := &IBeaconPeripheral{
			GenericPeripheral: &GenericPeripheral{},
		}

		err = mapToStruct(ibeacon, input)

		if err == nil {
			return ibeacon, nil
		}
	default:
		err = errors.Errorf("unsupported peripheral kind: %s", kind)
	}

	return nil, err
}

func FromPeripheral(target *tracking.Peripheral) (Peripheral, error) {
	var err error
	genericDto := &GenericPeripheral{
		Id:      target.Id,
		Kind:    target.Kind,
		Name:    target.Name,
		Enabled: target.Enabled,
	}

	switch target.Kind {
	case peripherals.PERIPHERAL_IBEACON:
		uuid, major, minor, err := peripherals.ParseIBeaconUniqueKey(target.Key)

		if err == nil {
			return &IBeaconPeripheral{
				GenericPeripheral: genericDto,
				Uuid:              uuid,
				Major:             major,
				Minor:             minor,
			}, nil
		}
	default:
		err = errors.Errorf("unsupported peripheral kind: '%s'", target.Kind)
	}

	return nil, err
}

func mapToStruct(target interface{}, values map[string]interface{}) error {
	for key, value := range values {
		err := utils.SetStructField(target, strings.Title(key), value)

		if err != nil {
			return err
		}
	}

	return nil
}
