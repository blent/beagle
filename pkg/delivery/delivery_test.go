package delivery_test

import (
	"github.com/blent/beagle/pkg/delivery"
	"github.com/blent/beagle/pkg/discovery/peripherals"
	"github.com/blent/beagle/pkg/notification"
	"github.com/brianvoe/gofakeit"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"testing"
	"time"
)

func TestSenderSingleSubscriber(t *testing.T) {
	sub := &notification.Subscriber{
		Id:    gofakeit.Uint64(),
		Name:  gofakeit.Username(),
		Event: notification.FOUND,
		Endpoint: &notification.Endpoint{
			Id:     gofakeit.Uint64(),
			Name:   gofakeit.Username(),
			Url:    gofakeit.URL(),
			Method: http.MethodPost,
		},
		Enabled: true,
	}

	resolver := func(req *http.Request) error {
		assert.Equal(t, sub.Endpoint.Url, req.URL.String(), "req url")

		return nil
	}

	logger := zap.NewNop()
	sender := delivery.New(logger, delivery.NewMockTransport(resolver))

	var notificationErr error

	sender.AddEventListener(func(evt delivery.Event) {
		notificationErr = evt.Error
	})

	err := sender.Send(notification.NewMessage(
		notification.FOUND,
		"test",
		createPeripheral(),
		[]*notification.Subscriber{sub},
	))

	assert.NoError(t, err, "send error")

	time.Sleep(time.Second + 5)

	assert.NoError(t, notificationErr, "delivery error")
}

func TestSenderMultipleSubscribers(t *testing.T) {
	max := gofakeit.Number(1, 10)
	subs := make([]*notification.Subscriber, 0, max)
	urls := make(map[string]string)

	for i := 0; i < max; i++ {
		url := gofakeit.URL()
		endpointName := gofakeit.Username()

		_, has := urls[url]

		if has == false {
			urls[url] = endpointName
		}

		subs = append(subs, &notification.Subscriber{
			Id:    uint64(i + 1),
			Name:  gofakeit.Username(),
			Event: notification.FOUND,
			Endpoint: &notification.Endpoint{
				Id:     uint64(i + 1),
				Name:   endpointName,
				Url:    url,
				Method: http.MethodPost,
			},
			Enabled: true,
		})
	}

	resolver := func(req *http.Request) error {
		_, has := urls[req.URL.String()]

		assert.True(t, has, "existing url")

		return nil
	}

	logger := zap.NewNop()
	sender := delivery.New(logger, delivery.NewMockTransport(resolver))

	var notificationErr error
	var counter int

	sender.AddEventListener(func(evt delivery.Event) {
		counter++

		if evt.Error != nil {
			notificationErr = evt.Error
		}
	})

	err := sender.Send(notification.NewMessage(
		notification.FOUND,
		"test",
		createPeripheral(),
		subs,
	))

	assert.NoError(t, err, "send error")

	time.Sleep(time.Second + 5)

	assert.Equal(t, counter, max, "dispatches")

	assert.NoError(t, notificationErr, "delivery error")
}

func TestSenderHandleFailure(t *testing.T) {
	sub := &notification.Subscriber{
		Id:    gofakeit.Uint64(),
		Name:  gofakeit.Username(),
		Event: notification.FOUND,
		Endpoint: &notification.Endpoint{
			Id:     gofakeit.Uint64(),
			Name:   gofakeit.Username(),
			Url:    gofakeit.URL(),
			Method: http.MethodPost,
		},
		Enabled: true,
	}

	resolver := func(req *http.Request) error {
		return errors.New("test error")
	}

	logger := zap.NewNop()
	sender := delivery.New(logger, delivery.NewMockTransport(resolver))

	var notificationErr error

	sender.AddEventListener(func(evt delivery.Event) {
		if evt.Error != nil {
			notificationErr = evt.Error
		}
	})

	err := sender.Send(notification.NewMessage(
		notification.FOUND,
		"test",
		createPeripheral(),
		[]*notification.Subscriber{sub},
	))

	assert.NoError(t, err, "send error")

	time.Sleep(time.Second + 5)

	assert.Error(t, notificationErr, "must be delivery error")
}

func createPeripheral() peripherals.Peripheral {
	return peripherals.NewMockPeripheral(
		gofakeit.UUID(),
		"mock",
		gofakeit.BuzzWord(),
		[]byte(gofakeit.HipsterSentence(5)),
		gofakeit.Float64(),
		gofakeit.Float64(),
		gofakeit.IPv4Address(),
	)
}
