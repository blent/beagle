package http

import (
	"fmt"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"net/http"
	"path/filepath"
	"strings"
)

type (
	Route interface {
		Use(gin.IRoutes)
	}

	Server struct {
		logger   *zap.Logger
		engine   *gin.Engine
		settings *Settings
	}
)

func NewServer(logger *zap.Logger, settings *Settings) *Server {
	if !settings.Enabled {
		return &Server{logger, nil, settings}
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(LoggerMiddleware(logger))

	return &Server{logger, engine, settings}
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
		dir, err := filepath.Abs(server.settings.Static.Directory)

		if err != nil {
			return err
		}

		server.engine.Use(static.Serve(
			server.settings.Static.Route,
			static.LocalFile(dir, true),
		))

		server.engine.NoRoute(func(ctx *gin.Context) {
			if strings.HasPrefix(ctx.Request.URL.Path, server.settings.Api.Route) {
				ctx.AbortWithError(http.StatusNotFound, errors.New("Route not found"))
				return
			}

			ctx.File(filepath.Join(dir, "index.html"))
		})
	}

	return server.engine.Run(fmt.Sprintf(":%d", server.settings.Port))
}
