package delivery

import (
	"net/http"
)

type Transport interface {
	Do(*http.Request) error
}
