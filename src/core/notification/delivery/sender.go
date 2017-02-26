package delivery

import (
	"fmt"
	"github.com/blent/beagle/src/core/discovery/peripherals"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/notification/delivery/transports"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/valyala/fasthttp"
	"net/http"
	"net/url"
	"strconv"
)

type (
	EventHandler func(targetName string, subscriber *tracking.Subscriber)

	Sender struct {
		logger    *logging.Logger
		transport transports.Transport
		handlers  map[string][]EventHandler
	}
)

func NewSender(logger *logging.Logger, transport transports.Transport) *Sender {
	return &Sender{
		logger,
		transport,
		make(map[string][]EventHandler),
	}
}

func (sender *Sender) Send(event *Event) error {
	if !sender.isSupportedEventName(event.Name()) {
		return fmt.Errorf("%s %s", ErrUnsupportedEventName, event.Name())
	}

	// Run bulk notification in a separate goroutine
	go sender.sendBatch(event)

	return nil
}

func (sender *Sender) Subscribe(eventName string, handler EventHandler) {
	if handler == nil {
		return
	}

	event := sender.handlers[eventName]

	if event == nil {
		event = make([]EventHandler, 0, 10)
	}

	sender.handlers[eventName] = append(event, handler)
}

func (sender *Sender) isSupportedEventName(name string) bool {
	if name == "" {
		return false
	}

	return name == "found" || name == "lost"
}

func (sender *Sender) sendBatch(event *Event) {
	succeeded := make([]*tracking.Subscriber, 0, len(event.Subscribers()))
	failed := make([]*tracking.Subscriber, 0, len(event.Subscribers()))

	for _, subscriber := range event.Subscribers() {
		err := sender.sendSingle(subscriber, event.Peripheral())

		if err == nil {
			succeeded = append(succeeded, subscriber)
		} else {
			failed = append(failed, subscriber)
		}
	}

	sender.emit("success", event.TargetName(), succeeded)
	sender.emit("failure", event.TargetName(), failed)
}

func (sender *Sender) sendSingle(subscriber *tracking.Subscriber, peripheral peripherals.Peripheral) error {
	serialized, err := sender.serializePeripheral(peripheral)

	if err != nil {
		sender.logger.Error(err.Error())
		return err
	}

	req := &fasthttp.Request{}

	if subscriber.Url == "" {
		err = fmt.Errorf("Subscriber %s has an empty url", subscriber.Name)
		sender.logger.Error(err.Error())
		return err
	}

	req.SetRequestURI(subscriber.Url)

	switch subscriber.Method {
	case http.MethodPost:
		req.SetBodyString(serialized.Encode())
	case http.MethodGet:
		uri := req.URI()
		uri.SetQueryString(serialized.Encode())
	}

	if req == nil {
		err = fmt.Errorf(
			"%s: %s for subscriber %s",
			ErrUnsupportedHttpMethod,
			subscriber.Method,
			subscriber.Name,
		)

		sender.logger.Errorf(err.Error())

		return err
	}

	headers := subscriber.Headers

	if headers != nil && len(headers) > 0 {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	err = sender.transport.Do(req)

	if err != nil {
		sender.logger.Errorf("Failed to notify subscriber %s", subscriber.Name)
		return err
	}

	return nil
}

func (sender *Sender) serializePeripheral(peripheral peripherals.Peripheral) (*url.Values, error) {
	serialized := &url.Values{}

	switch peripheral.Kind() {
	case peripherals.PERIPHERAL_IBEACON:
		ibeacon, ok := peripheral.(*peripherals.IBeaconPeripheral)

		if !ok {
			return nil, fmt.Errorf("%s %s", ErrUnabledToSerializePeripheral, peripheral.UniqueKey())
		}

		serialized.Set("uuid", ibeacon.Uuid())
		serialized.Set("major", strconv.Itoa(int(ibeacon.Major())))
		serialized.Set("minor", strconv.Itoa(int(ibeacon.Minor())))

	}

	serialized.Set("localName", peripheral.LocalName())
	serialized.Set("kind", peripheral.Kind())
	serialized.Set("proximity", peripheral.Proximity())
	serialized.Set("accuracy", strconv.FormatFloat(peripheral.Accuracy(), 'f', 6, 64))

	return serialized, nil
}

func (sender *Sender) emit(eventName, targetName string, subscribers []*tracking.Subscriber) {
	if subscribers == nil || len(subscribers) == 0 {
		return
	}

	event := sender.handlers[eventName]

	if event == nil {
		return
	}

	for _, handler := range event {
		for _, sub := range subscribers {
			handler(targetName, sub)
		}
	}
}
