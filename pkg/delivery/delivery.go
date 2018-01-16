package delivery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/blent/beagle/pkg/discovery/peripherals"
	"github.com/blent/beagle/pkg/notification"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type (
	Event struct {
		Timestamp  time.Time
		TargetName string
		Subscriber *notification.Subscriber
	}

	SuccessEvent struct {
		*Event
	}

	FailureEvent struct {
		*Event
		Reason error
	}

	SuccessListener func(evt SuccessEvent)

	FailureListener func(evt FailureEvent)

	Sender struct {
		logger    *zap.Logger
		transport Transport
		onSuccess []SuccessListener
		onFailure []FailureListener
	}
)

func New(logger *zap.Logger, transport Transport) *Sender {
	return &Sender{
		logger,
		transport,
		make([]SuccessListener, 0, 5),
		make([]FailureListener, 0, 5),
	}
}

func (sender *Sender) Send(msg *notification.Message) error {
	if !sender.isSupportedEventName(msg.EventName()) {
		return fmt.Errorf("%s %s", ErrUnsupportedEventName, msg.EventName())
	}

	// Call endpoints in batch inside a separate goroutine
	go sender.sendBatch(msg)

	return nil
}

func (sender *Sender) AddSuccessListener(listener SuccessListener) {
	if listener == nil {
		return
	}

	sender.onSuccess = append(sender.onSuccess, listener)
}

func (sender *Sender) RemoveSuccessListener(listener SuccessListener) bool {
	if listener == nil {
		return false
	}

	idx := -1
	handlerPointer := reflect.ValueOf(listener).Pointer()

	for i, element := range sender.onSuccess {
		currentPointer := reflect.ValueOf(element).Pointer()

		if currentPointer == handlerPointer {
			idx = i
		}
	}

	if idx < 0 {
		return false
	}

	sender.onSuccess = append(sender.onSuccess[:idx], sender.onSuccess[idx+1:]...)

	return true
}

func (sender *Sender) AddFailureListener(listener FailureListener) {
	if listener == nil {
		return
	}

	sender.onFailure = append(sender.onFailure, listener)
}

func (sender *Sender) RemoveFailureListener(listener FailureListener) bool {
	if listener == nil {
		return false
	}

	idx := -1
	handlerPointer := reflect.ValueOf(listener).Pointer()

	for i, element := range sender.onFailure {
		currentPointer := reflect.ValueOf(element).Pointer()

		if currentPointer == handlerPointer {
			idx = i
		}
	}

	if idx < 0 {
		return false
	}

	sender.onFailure = append(sender.onFailure[:idx], sender.onFailure[idx+1:]...)

	return true
}

func (sender *Sender) isSupportedEventName(name string) bool {
	if name == "" {
		return false
	}

	return name == "found" || name == "lost"
}

func (sender *Sender) sendBatch(msg *notification.Message) {
	subscribers := msg.Subscribers()
	succeeded := make([]*SuccessEvent, 0, len(subscribers))
	failed := make([]*FailureEvent, 0, len(subscribers))

	for _, subscriber := range subscribers {
		err := sender.sendSingle(msg.TargetName(), msg.Peripheral(), subscriber)

		evt := &Event{
			Timestamp:  time.Now(),
			TargetName: msg.TargetName(),
			Subscriber: subscriber,
		}

		if err == nil {
			sender.logger.Info(
				"Succeeded to notify a subscriber for peripheral",
				zap.String("subscriber", subscriber.Name),
				zap.String("peripheral", msg.TargetName()),
			)

			succeeded = append(succeeded, &SuccessEvent{evt})
		} else {
			sender.logger.Info(
				"Failed to notify a subscriber '%s' for peripheral '%s'",
				zap.String("subscriber", subscriber.Name),
				zap.String("peripheral", msg.TargetName()),
				zap.Error(err),
			)

			failed = append(failed, &FailureEvent{evt, err})
		}
	}

	sender.emitSuccess(succeeded)
	sender.emitFailure(failed)
}

func (sender *Sender) sendSingle(name string, peripheral peripherals.Peripheral, subscriber *notification.Subscriber) error {
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
	if peripheral == nil {
		return nil, errors.New("missed peripheral")
	}

	serialized := make(map[string]interface{})

	serialized["name"] = name
	serialized["kind"] = peripheral.Kind()
	serialized["proximity"] = peripheral.Proximity()
	serialized["accuracy"] = strconv.FormatFloat(peripheral.Accuracy(), 'f', 6, 64)

	switch peripheral.Kind() {
	case peripherals.PERIPHERAL_IBEACON:
		ibeacon, ok := peripheral.(*peripherals.IBeaconPeripheral)

		if !ok {
			return nil, fmt.Errorf("%s %s", ErrUnableToSerializePeripheral, peripheral.UniqueKey())
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

func (sender *Sender) emitSuccess(events []*SuccessEvent) {
	if events == nil || len(events) == 0 {
		return
	}

	for _, listener := range sender.onSuccess {
		for _, evt := range events {
			listener(*evt)
		}
	}
}

func (sender *Sender) emitFailure(events []*FailureEvent) {
	if events == nil || len(events) == 0 {
		return
	}

	for _, listener := range sender.onFailure {
		for _, evt := range events {
			listener(*evt)
		}
	}
}
