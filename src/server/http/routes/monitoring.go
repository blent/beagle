package routes

import (
	"github.com/blent/beagle/src/server/monitoring/activity"
	"github.com/blent/beagle/src/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"path"
)

type MonitoringRoute struct {
	baseUrl  string
	logger   *zap.Logger
	activity *activity.Service
}

func NewMonitoringRoute(baseUrl string, logger *zap.Logger, activity *activity.Service) *MonitoringRoute {
	return &MonitoringRoute{baseUrl, logger, activity}
}

func (rt *MonitoringRoute) Use(routes gin.IRoutes) {
	routes.GET(path.Join("/", rt.baseUrl, "activity"), func(ctx *gin.Context) {
		take, err := utils.StringToUint64(ctx.Query("take"))

		if err != nil {
			rt.logger.Error("failed to parse parameter: take")
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid parameter: take"))
			return
		}

		skip, err := utils.StringToUint64(ctx.Query("skip"))

		if err != nil {
			rt.logger.Error("failed to parse parameter: skip")
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid parameter: skip"))
			return
		}

		ctx.JSON(200, gin.H{
			"items":    rt.activity.GetRecords(int(take), int(skip)),
			"quantity": rt.activity.Quantity(),
		})
	})
}
