package routes

import (
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/server/history/activity"
	"github.com/gin-gonic/gin"
	"path"
)

type ActivityRoute struct {
	baseEndpoint string
	logger       *logging.Logger
	writer       *activity.Writer
}

func NewActivityRoute(baseEndpoint string, logger *logging.Logger, writer *activity.Writer) *ActivityRoute {
	return &ActivityRoute{baseEndpoint, logger, writer}
}

func (rt *ActivityRoute) Use(routes gin.IRoutes) {
	route := "activity"

	routes.GET(path.Join("/", rt.baseEndpoint, route), func(ctx *gin.Context) {
		ctx.JSON(200, rt.writer.GetCurrent())
	})
}
