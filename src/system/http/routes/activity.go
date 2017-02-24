package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/system/history/activity"
	"path"
)

type ActivityRoute struct {
	baseEndpoint string
	logger       *logging.Logger
	writer       *activity.Writer
}

func NewActivityRoutes(baseEndpoint string, logger *logging.Logger, writer *activity.Writer) *ActivityRoute {
	return &ActivityRoute{baseEndpoint, logger, writer}
}

func (rt *ActivityRoute) Use(routes gin.IRoutes) {
	route := "activity"

	routes.GET(path.Join("/", rt.baseEndpoint, route), func(ctx *gin.Context) {
		ctx.JSON(200, rt.writer.GetCurrent())
	})
}
