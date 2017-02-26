package routes

import (
	"github.com/blent/beagle/src/core/logging"
	"github.com/gin-gonic/gin"
)

type StaticsRoute struct {
	logger *logging.Logger
}

func NewStaticsRoute(logger *logging.Logger) *StaticsRoute {
	return &StaticsRoute{logger}
}

func (rt StaticsRoute) Use(routes gin.IRoutes) {
}
