package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/blent/beagle/src/core/discovery/peripherals"
	"github.com/blent/beagle/src/core/notification/transport"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type (
	SenderEventHandler func(targetName string, subscriber *Subscriber)

	Sender struct {
		logger    *zap.Logger
		transport transport.Transport
		handlers  map[string][]SenderEventHandler
	}
)

func NewSender(logger *zap.Logger, transport transport.Transport) *Sender {
	return &Sender{
		logger,
		transport,
		make(map[string][]SenderEventHandler),
	}
}

func (sender *Sender) Send(msg *Message) error {
	if !sender.isSupportedEventName(msg.EventName()) {
		return fmt.Errorf("%s %s", ErrUnsupportedEventName, msg.EventName())
	}

	// Call endpoints in batch inside a separate goroutine
	go sender.sendBatch(msg)

	return nil
}

func (sender *Sender) Subscribe(eventName string, handler SenderEventHandler) {
	if handler == nil {
		return
	}

	event := sender.handlers[eventName]

	if event == nil {
		event = make([]SenderEventHandler, 0, 10)
	}

	sender.handlers[eventName] = append(event, handler)
}

func (sender *Sender) isSupportedEventName(name string) bool {
	if name == "" {
		return false
	}

	return name == "found" || name == "lost"
}

func (sender *Sender) sendBatch(msg *Message) {
	subscribers := msg.Subscribers()
	succeeded := make([]*Subscriber, 0, len(subscribers))
	failed := make([]*Subscriber, 0, len(subscribers))

	for _, subscriber := range subscribers {
		err := sender.sendSingle(msg.TargetName(), msg.Peripheral(), subscriber)

		if err == nil {
			succeeded = append(succeeded, subscriber)

			sender.logger.Info(
				"Succeeded to notify a subscriber for peripheral",
				zap.String("subscriber", subscriber.Name),
				zap.String("peripheral", msg.TargetName()),
			)
		} else {
			failed = append(failed, subscriber)

			sender.logger.Info(
				"Failed to notify a subscriber '%s' for peripheral '%s'",
				zap.String("subscriber", subscriber.Name),
				zap.String("peripheral", msg.TargetName()),
				zap.Error(err),
			)
		}
	}

	sender.emit("success", msg.TargetName(), succeeded)
	sender.emit("failure", msg.TargetName(), failed)
}

func (sender *Sender) sendSingle(name string, peripheral peripherals.Peripheral, subscriber *Subscriber) error {
	serialized, err := sender.serializePeripheral(name, peripheral)

	if err != nil {
		sender.logger.Error(err.Error())
		return err
	}

	endpoint := subscriber.Endpoint

	if endpoint == nil {
		sender.logger.Warn(
			"subscriber has no endpoints",
			zap.String("subscriber", subscriber.Name),
		)
		return nil
	}

	if endpoint.Url == "" {
		err = errors.New("Endpoint has an empty url")

		sender.logger.Error(
			"endpoint has an empty url: %s",
			zap.String("endpoint", endpoint.Name),
			zap.Error(err),
		)

		return err
	}

	method := strings.ToUpper(endpoint.Method)
	req, err := http.NewRequest(method, subscriber.Endpoint.Url, nil)

	if err != nil {
		sender.logger.Error(
			"failed to create a new request",
			zap.Error(err),
			zap.String("endpoint", endpoint.Name),
		)

		return errors.Wrap(err, "failed to create a new request")
	}

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")

		body, err := json.Marshal(serialized)

		if err != nil {
			return err
		}

		req.Body = ioutil.NopCloser(bytes.NewReader(body))
	} else {
		query, err := sender.encode(serialized)

		if err != nil {
			return err
		}

		req.URL.RawQuery = query
	}

	if req == nil {
		err = fmt.Errorf(
			"%s: %s for endpoint %s",
			ErrUnsupportedHttpMethod,
			endpoint.Method,
			endpoint.Name,
		)

		sender.logger.Error(
			"Failed to create a request",
			zap.String("endpoint", endpoint.Name),
			zap.Error(err),
		)

		return err
	}

	headers := endpoint.Headers

	if headers != nil && len(headers) > 0 {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	err = sender.transport.Do(req)

	if err != nil {
		sender.logger.Error(
			"Failed to reach out the endpoint",
			zap.String("endpoint name", endpoint.Name),
			zap.String("endpoint url", endpoint.Url),
			zap.Error(err),
		)

		return err
	}

	return nil
}

func (sender *Sender) serializePeripheral(name string, peripheral peripherals.Peripheral) (map[string]interface{}, error) {
	serialized := make(map[string]interface{})

	serialized["name"] = name
	serialized["kind"] = peripheral.Kind()
	serialized["proximity"] = peripheral.Proximity()
	serialized["accuracy"] = strconv.FormatFloat(peripheral.Accuracy(), 'f', 6, 64)

	switch peripheral.Kind() {
	case peripherals.PERIPHERAL_IBEACON:
		ibeacon, ok := peripheral.(*peripherals.IBeaconPeripheral)

		if !ok {
			return nil, fmt.Errorf("%s %s", ErrUnabledToSerializePeripheral, peripheral.UniqueKey())
		}

		serialized["uuid"] = ibeacon.Uuid()
		serialized["major"] = strconv.Itoa(int(ibeacon.Major()))
		serialized["minor"] = strconv.Itoa(int(ibeacon.Minor()))
	}

	return serialized, nil
}

func (sender *Sender) encode(data map[string]interface{}) (string, error) {
	var buf bytes.Buffer

	for k, v := range data {
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(fmt.Sprintf("%s", v))
		buf.WriteByte('&')
	}

	str := buf.String()

	// remove last ampersand
	return str[0 : len(str)-1], nil
}

func (sender *Sender) emit(eventName, targetName string, subscribers []*Subscriber) {
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
