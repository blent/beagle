package notification

import (
	"fmt"
	"github.com/blent/beagle/src/core/discovery/peripherals"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/notification/transports"
	"github.com/valyala/fasthttp"
	"net/http"
	"net/url"
	"strconv"
	"encoding/json"
	"strings"
)

type (
	SenderEventHandler func(targetName string, subscriber *Subscriber)

	Sender struct {
		logger    *logging.Logger
		transport transports.Transport
		handlers  map[string][]SenderEventHandler
	}
)

func NewSender(logger *logging.Logger, transport transports.Transport) *Sender {
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
			sender.logger.Infof(
				"Succeeded to notify a subscriber '%s' for peripheral '%s'",
				subscriber.Name,
				msg.TargetName(),
			)
		} else {
			failed = append(failed, subscriber)
			sender.logger.Infof(
				"Failed to notify a subscriber '%s' for peripheral '%s'",
				subscriber.Name,
				msg.TargetName(),
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
		sender.logger.Warnf("Subscriber has no endpoints: %s", subscriber.Name)
		return nil
	}

	if endpoint.Url == "" {
		err = fmt.Errorf("Endpoint has an empty url: %s", endpoint.Name)
		sender.logger.Error(err.Error())
		return err
	}

	method := strings.ToUpper(endpoint.Method)
	req := &fasthttp.Request{}
	req.Header.SetMethod(method)
	req.SetRequestURI(subscriber.Endpoint.Url)

	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")

		body, err := json.Marshal(serialized)

		if err != nil {
			return err
		}

		req.SetBody(body)
	} else {
		req.URI().SetQueryString(serialized.Encode())
	}

	sender.logger.Infof("Target url is %s", req.URI().String())

	if req == nil {
		err = fmt.Errorf(
			"%s: %s for endpoint %s",
			ErrUnsupportedHttpMethod,
			endpoint.Method,
			endpoint.Name,
		)

		sender.logger.Errorf(err.Error())

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
		sender.logger.Errorf("Failed to reach endpoint %s", endpoint.Name)
		return err
	}

	return nil
}

func (sender *Sender) serializePeripheral(name string, peripheral peripherals.Peripheral) (*url.Values, error) {
	serialized := &url.Values{}

	serialized.Set("name", name)
	serialized.Set("kind", peripheral.Kind())
	serialized.Set("proximity", peripheral.Proximity())
	serialized.Set("accuracy", strconv.FormatFloat(peripheral.Accuracy(), 'f', 6, 64))

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

	return serialized, nil
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
