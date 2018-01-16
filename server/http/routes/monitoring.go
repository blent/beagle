package routes

import (
	"github.com/blent/beagle/pkg/monitoring/activity"
	"github.com/blent/beagle/pkg/monitoring/system"
	"github.com/blent/beagle/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"path"
)

type MonitoringRoute struct {
	baseUrl  string
	logger   *zap.Logger
	activity *activity.Monitoring
	system   *system.Monitoring
}

func NewMonitoringRoute(baseUrl string, logger *zap.Logger, activity *activity.Monitoring, system *system.Monitoring) *MonitoringRoute {
	return &MonitoringRoute{baseUrl, logger, activity, system}
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

		ctx.JSON(http.StatusOK, gin.H{
			"items":    rt.activity.GetRecords(int(take), int(skip)),
			"quantity": rt.activity.Quantity(),
		})
	})

	routes.GET(path.Join("/", rt.baseUrl, "system"), func(ctx *gin.Context) {
		stats, err := rt.system.GetStats()

		if err != nil {
			rt.logger.Error(
				"Failed to retrieve system stats",
				zap.Error(err),
			)

			ctx.AbortWithStatus(http.StatusInternalServerError)

			return
		}

		ctx.JSON(http.StatusOK, stats)
	})
}
