package notification

import (
	"github.com/blent/beagle/pkg/discovery/peripherals"
	"github.com/blent/beagle/pkg/tracking"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"reflect"
	"time"
)

type (
	Event struct {
		Name       string                 `json:"name"`
		Timestamp  time.Time              `json:"timestamp"`
		Peripheral peripherals.Peripheral `json:"peripheral"`
		Registered bool                   `json:"registered"`
	}

	EventListener func(evt Event)

	Registry interface {
		FindTarget(key string) (*tracking.Peripheral, error)

		FindSubscribers(targetId uint64, events ...string) ([]*Subscriber, error)
	}

	MessageSender interface {
		Send(msg *Message) error
	}

	Broker struct {
		logger    *zap.Logger
		sender    MessageSender
		registry  Registry
		listeners []EventListener
	}
)

func NewBroker(logger *zap.Logger, sender MessageSender, registry Registry) (*Broker, error) {
	if logger == nil {
		return nil, errors.Wrap(ErrMissedArg, "logger")
	}

	if sender == nil {
		return nil, errors.Wrap(ErrMissedArg, "sender")
	}

	if registry == nil {
		return nil, errors.Wrap(ErrMissedArg, "registry")
	}

	return &Broker{
		logger,
		sender,
		registry,
		make([]EventListener, 0, 5),
	}, nil
}

func (broker *Broker) Use(stream *tracking.Stream) {
	go broker.doUse(stream)
}

func (broker *Broker) AddEventListener(listener EventListener) {
	if listener == nil {
		return
	}

	broker.listeners = append(broker.listeners, listener)
}

func (broker *Broker) RemoveEventListener(listener EventListener) bool {
	if listener == nil {
		return false
	}

	idx := -1
	handlerPointer := reflect.ValueOf(listener).Pointer()

	for i, element := range broker.listeners {
		currentPointer := reflect.ValueOf(element).Pointer()

		if currentPointer == handlerPointer {
			idx = i
		}
	}

	if idx < 0 {
		return false
	}

	broker.listeners = append(broker.listeners[:idx], broker.listeners[idx+1:]...)

	return true
}

func (broker *Broker) doUse(stream *tracking.Stream) {
	streamIsClosed := false

	for {
		if streamIsClosed {
			broker.logger.Info("Stream is closed")
			return
		}

		select {
		case peripheral, isOpen := <-stream.Found():
			if isOpen {
				broker.notify(FOUND, peripheral)
			}

			streamIsClosed = !isOpen
		case peripheral, isOpen := <-stream.Lost():
			if isOpen {
				broker.notify(LOST, peripheral)
			}

			streamIsClosed = !isOpen
		case err, _ := <-stream.Error():
			streamIsClosed = true

			broker.logger.Error(
				"Error occurred during consuming the stream",
				zap.Error(err),
			)
		}
	}
}

func (broker *Broker) notify(eventName string, peripheral peripherals.Peripheral) {
	go func() {
		key := peripheral.UniqueKey()

		if key == "" {
			broker.logger.Error("Peripheral contains an empty key")
			return
		}

		found, err := broker.registry.FindTarget(key)

		evt := &Event{
			Timestamp:  time.Now(),
			Name:       eventName,
			Peripheral: peripheral,
			Registered: found != nil,
		}

		if err != nil {
			broker.logger.Error(
				"Failed to retrieve a peripheral",
				zap.String("key", key),
				zap.Error(err),
			)

			broker.emit(evt)

			return
		}

		broker.emit(evt)

		if found == nil {
			broker.logger.Info(
				"Peripheral is not registered",
				zap.String("key", key),
			)

			return
		}

		if found.Enabled == false {
			broker.logger.Info(
				"Peripheral is disabled",
				zap.String("key", key),
			)

			return
		}

		subscribers, err := broker.registry.FindSubscribers(found.Id, eventName, "*")

		if subscribers == nil || len(subscribers) == 0 {
			broker.logger.Info(
				"Peripheral does not have any enabled subscribers",
				zap.String("key", key),
			)
			return
		}

		broker.sender.Send(NewMessage(eventName, found.Name, peripheral, subscribers))
	}()
}

func (broker *Broker) emit(evt *Event) {
	go func() {
		for _, handler := range broker.listeners {
			handler(*evt)
		}
	}()
}
