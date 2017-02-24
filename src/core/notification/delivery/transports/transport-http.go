package transports

import "github.com/valyala/fasthttp"

type HttpTransport struct {
	engine *fasthttp.Client
}

func NewHttpTransport() *HttpTransport {
	return &HttpTransport{
		engine: &fasthttp.Client{},
	}
}

func (transport *HttpTransport) Do(req *fasthttp.Request) error {
	return transport.engine.Do(req, nil)
}
