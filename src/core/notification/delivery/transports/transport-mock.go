package transports

import "github.com/valyala/fasthttp"

type (
	MockTransport struct {
		engine func(req *fasthttp.Request) error
	}
)

func NewMockTransport(engine func(req *fasthttp.Request) error) *MockTransport {
	return &MockTransport{engine}
}

func (transport *MockTransport) Do(req *fasthttp.Request) error {
	if transport.engine != nil {
		return transport.engine(req)
	}

	return nil
}
