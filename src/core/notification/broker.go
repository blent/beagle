package notification

import (
	"github.com/blent/beagle/src/core/discovery/peripherals"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/notification/delivery"
	"github.com/blent/beagle/src/core/tracking"
	"strings"
)

type (
	EventHandler func(target *tracking.Target, peripheral peripherals.Peripheral)

	Registry func(key string) (*tracking.Target, error)

	EventBroker struct {
		logger   *logging.Logger
		sender   *delivery.Sender
		registry Registry
		handlers map[string][]EventHandler
	}
)

func NewEventBroker(logger *logging.Logger, sender *delivery.Sender, registry Registry) *EventBroker {
	return &EventBroker{logger, sender, registry, make(map[string][]EventHandler)}
}

func (broker *EventBroker) Use(stream *tracking.Stream) {
	go broker.doUse(stream)
}

func (broker *EventBroker) Subscribe(eventName string, handler EventHandler) {
	if handler != nil {
		event := broker.handlers[eventName]

		if event == nil {
			event = make([]EventHandler, 0, 10)
		}

		broker.handlers[eventName] = append(event, handler)
	}
}

func (broker *EventBroker) doUse(stream *tracking.Stream) {
	streamIsClosed := false

	for {
		if streamIsClosed {
			broker.logger.Info("Stream is closed")
			return
		}

		select {
		case peripheral, isOpen := <-stream.Found():
			if isOpen {
				broker.notify(PERIPHERAL_FOUND, peripheral)
			}

			streamIsClosed = !isOpen
		case peripheral, isOpen := <-stream.Lost():
			if isOpen {
				broker.notify(PERIPHERAL_LOST, peripheral)
			}

			streamIsClosed = !isOpen
		case err, _ := <-stream.Error():
			streamIsClosed = true

			broker.logger.Errorf("Error occured during consuming the stream %s", err.Error())
		}
	}
}

func (broker *EventBroker) notify(eventName string, peripheral peripherals.Peripheral) {
	key := peripheral.UniqueKey()
	found, err := broker.registry(key)

	if err != nil {
		broker.logger.Errorf("Failed to retrieve target with key %s: %s", key, err.Error())
		return
	}

	if found == nil {
		broker.logger.Infof("Peripheral with key %s is not registered", key)
		return
	}

	if found.Subscribers == nil || len(found.Subscribers) == 0 {
		broker.logger.Infof("Peripheral with key %s does not have subscribers")
	}

	subscribers := make([]*tracking.Subscriber, 0, len(found.Subscribers))

	for _, sub := range found.Subscribers {
		if sub.Event == "*" || strings.ToLower(sub.Event) == eventName {
			subscribers = append(subscribers, sub)
		}
	}

	broker.sender.Send(delivery.NewEvent(eventName, found.Name, peripheral, subscribers))
	broker.emit(eventName, found, peripheral)
}

func (broker *EventBroker) emit(eventName string, target *tracking.Target, peripheral peripherals.Peripheral) {
	go func() {
		event := broker.handlers[eventName]

		if event != nil {
			for _, handler := range event {
				handler(target, peripheral)
			}
		}
	}()
}
