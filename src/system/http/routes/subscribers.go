package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/system/storage"
)

type SubscribersRoute struct {
	logger *logging.Logger
	repo   storage.TargetRepository
}

func NewSubscribersRoute(logger *logging.Logger, repo storage.TargetRepository) *SubscribersRoute {
	return &SubscribersRoute{}
}

func (rt SubscribersRoute) Use(routes gin.IRoutes) {
}
