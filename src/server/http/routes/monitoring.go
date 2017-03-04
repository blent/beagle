package routes

import (
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/server/monitoring/activity"
	"github.com/gin-gonic/gin"
	"path"
)

type MonitoringRoute struct {
	baseUrl  string
	logger   *logging.Logger
	activity *activity.Service
}

func NewMonitoringRoute(baseUrl string, logger *logging.Logger, activity *activity.Service) *MonitoringRoute {
	return &MonitoringRoute{baseUrl, logger, activity}
}

func (rt *MonitoringRoute) Use(routes gin.IRoutes) {
	route := "monitoring"

	routes.GET(path.Join("/", rt.baseUrl, route, "activity"), func(ctx *gin.Context) {
		ctx.JSON(200, rt.activity.GetRecords())
	})
}
