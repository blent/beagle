package routes

import (
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/server/storage"
	"github.com/gin-gonic/gin"
	"net/http"
	"path"
	"strconv"
	"github.com/pkg/errors"
	"fmt"
)

var (
	ErrTargetRouteInvalidTarget = errors.New("invalid target")
)

type TargetRoute struct {
	baseEndpoint string
	logger       *logging.Logger
	repo         storage.TargetRepository
}

func NewTargetRoute(baseEndpoint string, logger *logging.Logger, repo storage.TargetRepository) *TargetRoute {
	return &TargetRoute{baseEndpoint, logger, repo}
}

func (rt *TargetRoute) Use(routes gin.IRoutes) {
	route := "target"

	routes.GET(path.Join("/", rt.baseEndpoint, "targets"), rt.findTargets)
	routes.GET(path.Join("/", rt.baseEndpoint, "targets", ":take"), rt.findTargets)
	routes.GET(path.Join("/", rt.baseEndpoint, "targets", ":take", ":skip"), rt.findTargets)
	routes.GET(path.Join("/", rt.baseEndpoint, route, ":id"), rt.getTarget)
	routes.POST(path.Join("/", rt.baseEndpoint, route), rt.createTarget)
	routes.PUT(path.Join("/", rt.baseEndpoint, route, ":id"), rt.updateTarget)
	routes.DELETE(path.Join("/", rt.baseEndpoint, route, ":id"), rt.deleteTarget)
}

func (rt *TargetRoute) getTarget(ctx *gin.Context) {
	id, ok := rt.parseUintParam(ctx, "id", true)

	if !ok {
		return
	}

	target, err := rt.repo.GetById(id)

	if err != nil {
		rt.logger.Error(fmt.Sprintf("Failed to retrieve target with id %d: %s", id, err.Error()))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if target == nil {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}

	ctx.JSON(http.StatusOK, target)
}

func (rt *TargetRoute) findTargets(ctx *gin.Context) {
	take, ok := rt.parseUintParam(ctx, "take", false)

	if !ok {
		return
	}

	skip, ok := rt.parseUintParam(ctx, "skip", false)

	if !ok {
		return
	}

	targets, err := rt.repo.Find(storage.NewTargetQuery(take, skip, storage.TARGET_STATUS_ANY))

	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, targets)
}

func (rt *TargetRoute) createTarget(ctx *gin.Context) {
	//target, ok := rt.parseTarget(ctx)
	//
	//if !ok {
	//	return
	//}
	//
	//if err := rt.repo.Update(target); err != nil {
	//	ctx.AbortWithStatus(http.StatusInternalServerError)
	//	return
	//}
	//
	//ctx.String(http.StatusOK, "%n", target.ID)

	ctx.AbortWithStatus(http.StatusNoContent)
}

func (rt *TargetRoute) updateTarget(ctx *gin.Context) {
	//target, ok := rt.parseTarget(ctx)
	//
	//if !ok {
	//	return
	//}
	//
	//if err := rt.repo.Update(target); err != nil {
	//	ctx.AbortWithStatus(http.StatusInternalServerError)
	//	return
	//}

	ctx.AbortWithStatus(http.StatusNoContent)
}

func (rt *TargetRoute) deleteTarget(ctx *gin.Context) {
	//id, ok := rt.parseId(ctx)
	//
	//if !ok {
	//	return
	//}
	//
	//err := rt.repo.Delete(id)
	//
	//if err != nil {
	//	ctx.AbortWithStatus(http.StatusInternalServerError)
	//	return
	//}

	ctx.AbortWithStatus(http.StatusNoContent)
}

func (rt *TargetRoute) parseUintParam(ctx *gin.Context, name string, required bool) (uint, bool) {
	idStr := ctx.Params.ByName(name)

	if idStr == "" {
		if required {
			ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid target '%s' parameter", name))
			return 0, false
		}

		return 0, true
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid target '%s' parameter", name))
		return 0, false
	}

	return uint(id), true
}

func (rt *TargetRoute) parseTarget(ctx *gin.Context) (*tracking.Target, bool) {
	var target *tracking.Target

	if err := ctx.BindJSON(&target); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, ErrTargetRouteInvalidTarget)
		return nil, false
	}

	return target, true
}
