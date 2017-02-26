package initializers

import (
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/server/http"
)

type RoutesInitializer struct {
	logger *logging.Logger
	server *http.Server
	routes []http.Route
}

func NewRoutesInitializer(logger *logging.Logger, server *http.Server, routes []http.Route) *RoutesInitializer {
	return &RoutesInitializer{logger, server, routes}
}

func (init *RoutesInitializer) Run() error {
	for _, route := range init.routes {
		init.server.AddRoute(route)
	}

	return nil
}
