package routes

import (
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/server/storage"
	"github.com/gin-gonic/gin"
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
