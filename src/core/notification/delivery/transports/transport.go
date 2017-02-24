package transports

import "github.com/valyala/fasthttp"

type Transport interface {
	Do(*fasthttp.Request) error
}
