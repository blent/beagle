package http

import (
	"fmt"
	"github.com/blent/beagle/src/core/logging"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

type (
	Route interface {
		Use(gin.IRoutes)
	}

	Server struct {
		logger   *logging.Logger
		engine   *gin.Engine
		settings *Settings
	}
)

func NewServer(logger *logging.Logger, settings *Settings) *Server {
	if !settings.Enabled {
		return &Server{logger, nil, settings}
	}

	return &Server{logger, gin.Default(), settings}
}

func (server *Server) AddRoute(route Route) *Server {
	if route == nil {
		return server
	}

	if server.engine != nil {
		route.Use(server.engine)
	}

	return server
}

func (server *Server) Run(ctx context.Context) error {
	if server.engine == nil {
		server.logger.Info("Server is disabled")
		<-ctx.Done()
		return nil
	}

	return server.engine.Run(fmt.Sprintf(":%d", server.settings.Port))
}
