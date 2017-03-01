package routes

import (
	"fmt"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/server/http/routes/dto"
	"github.com/blent/beagle/src/server/storage"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"path"
	"strconv"
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

	// Get multiple targets
	routes.GET(path.Join("/", rt.baseEndpoint, "targets"), rt.findTargets)
	routes.GET(path.Join("/", rt.baseEndpoint, "targets", ":take"), rt.findTargets)
	routes.GET(path.Join("/", rt.baseEndpoint, "targets", ":take", ":skip"), rt.findTargets)

	// Get single target by id
	routes.GET(path.Join("/", rt.baseEndpoint, route, ":id"), rt.getTarget)

	// Create new target
	routes.POST(path.Join("/", rt.baseEndpoint, route), rt.createTarget)

	// Update existing target by id
	routes.PUT(path.Join("/", rt.baseEndpoint, route, ":id"), rt.updateTarget)

	// Delete existing target by id
	routes.DELETE(path.Join("/", rt.baseEndpoint, route, ":id"), rt.deleteTarget)
}

func (rt *TargetRoute) getTarget(ctx *gin.Context) {
	id, ok := rt.parseParamUint64(ctx, "id", true)

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

	targetDto, ok := rt.serializeTarget(ctx, target)

	if !ok {
		return
	}

	ctx.JSON(http.StatusOK, targetDto)
}

func (rt *TargetRoute) findTargets(ctx *gin.Context) {
	take, ok := rt.parseParamUint64(ctx, "take", false)

	if !ok {
		return
	}

	skip, ok := rt.parseParamUint64(ctx, "skip", false)

	if !ok {
		return
	}

	targets, err := rt.repo.Find(storage.NewTargetQuery(take, skip, storage.TARGET_STATUS_ANY))

	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	targetsDto := make([]*dto.Target, 0, len(targets))

	for _, target := range targets {
		targetDto, ok := rt.serializeTarget(ctx, target)

		if !ok {
			return
		}

		targetsDto = append(targetsDto, targetDto)
	}

	ctx.JSON(http.StatusOK, targetsDto)
}

func (rt *TargetRoute) createTarget(ctx *gin.Context) {
	target, ok := rt.deserializeTarget(ctx)

	if !ok {
		return
	}

	id, err := rt.repo.Create(target)

	if err != nil {
		rt.logger.Errorf("Failed to create new target: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.String(http.StatusOK, "%d", id)
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

func (rt *TargetRoute) parseParamUint64(ctx *gin.Context, name string, required bool) (uint64, bool) {
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

	return id, true
}

func (rt *TargetRoute) serializeTarget(ctx *gin.Context, target *tracking.Target) (*dto.Target, bool) {
	targetDto, err := dto.FromTarget(target)

	if err != nil {
		rt.logger.Errorf("Failed to serialize target: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrTargetRouteInvalidTarget)

		return nil, false
	}

	return targetDto, true
}

func (rt *TargetRoute) deserializeTarget(ctx *gin.Context) (*tracking.Target, bool) {
	var targetDto *dto.Target

	err := ctx.BindJSON(&targetDto)

	if err != nil {
		rt.logger.Errorf("Failed to deserialize target: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrTargetRouteInvalidTarget)

		return nil, false
	}

	target, err := dto.ToTarget(targetDto)

	if err != nil {
		rt.logger.Errorf("Failed to deserialize target: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrTargetRouteInvalidTarget)

		return nil, false
	}

	return target, true
}
