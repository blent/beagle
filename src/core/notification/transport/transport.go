package transport

import (
	"net/http"
)

type Transport interface {
	Do(*http.Request) error
}
