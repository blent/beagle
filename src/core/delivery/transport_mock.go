package delivery

import (
	"net/http"
)

type (
	MockTransport struct {
		engine func(req *http.Request) error
	}
)

func NewMockTransport(engine func(req *http.Request) error) *MockTransport {
	return &MockTransport{engine}
}

func (transport *MockTransport) Do(req *http.Request) error {
	if transport.engine != nil {
		return transport.engine(req)
	}

	return nil
}
