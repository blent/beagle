package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/blent/beagle/src/core/logging"
)

type StaticsRoute struct {
	logger *logging.Logger
}

func NewStaticsRoute(logger *logging.Logger) *StaticsRoute {
	return &StaticsRoute{logger}
}

func (rt StaticsRoute) Use(routes gin.IRoutes) {
}
