package initializers

import (
	"github.com/blent/beagle/server/http"
	"go.uber.org/zap"
)

type RoutesInitializer struct {
	logger *zap.Logger
	server *http.Server
	routes []http.Route
}

func NewRoutesInitializer(logger *zap.Logger, server *http.Server, routes []http.Route) *RoutesInitializer {
	return &RoutesInitializer{logger, server, routes}
}

func (init *RoutesInitializer) Run() error {
	for _, route := range init.routes {
		init.server.AddRoute(route)
	}

	return nil
}
