package http

import (
	"fmt"
	"github.com/blent/beagle/src/core/logging"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"path/filepath"
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

	if server.settings.Static != nil {
		if server.settings.Static.Route != "" && server.settings.Static.Directory != "" {
			dir, err := filepath.Abs(server.settings.Static.Directory)

			if err != nil {
				return err
			}

			server.engine.Static(
				server.settings.Static.Route,
				dir,
			)

			server.engine.StaticFile(
				"/favicon.ico",
				filepath.Join(dir, "favicon.ico"),
			)

			server.engine.StaticFile(
				"/",
				filepath.Join(dir, "index.html"),
			)

			server.engine.NoRoute(func (ctx *gin.Context) {
				ctx.File(filepath.Join(dir, "index.html"))
			})
		}
	}

	return server.engine.Run(fmt.Sprintf(":%d", server.settings.Port))
}
