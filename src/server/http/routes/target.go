package routes

import (
	"fmt"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/server/http/routes/dto"
	"github.com/blent/beagle/src/server/storage"
	"github.com/blent/beagle/src/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"path"
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
	routes.GET(path.Join("/", rt.baseEndpoint, route), rt.findTargets)

	// Get single target by id
	routes.GET(path.Join("/", rt.baseEndpoint, route, ":id"), rt.getTarget)

	// Create new target
	routes.POST(path.Join("/", rt.baseEndpoint, route), rt.createTarget)

	// Update existing target by id
	routes.PUT(path.Join("/", rt.baseEndpoint, route), rt.updateTarget)

	// Delete existing target by id
	routes.DELETE(path.Join("/", rt.baseEndpoint, route), rt.deleteTarget)
}

func (rt *TargetRoute) getTarget(ctx *gin.Context) {
	id, err := utils.StringToUint64(ctx.Params.ByName("id"))

	if err != nil {
		rt.logger.Error(fmt.Sprintf("Failed to parse target id: %s", err.Error()))
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	target, err := rt.repo.GetById(id)

	if err != nil {
		rt.logger.Error(fmt.Sprintf("Failed to retrieve target with id %d: %s", id, err.Error()))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if target == nil {
		ctx.AbortWithStatus(http.StatusOK)
		return
	}

	targetDto, ok := rt.serializeTarget(ctx, target)

	if !ok {
		return
	}

	ctx.JSON(http.StatusOK, targetDto)
}

func (rt *TargetRoute) findTargets(ctx *gin.Context) {
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
	target, ok := rt.deserializeTarget(ctx)

	if !ok {
		return
	}

	if target.Id < 0 || target.Id == 0 {
		rt.logger.Error("Missed target id")
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	err := rt.repo.Update(target)

	if err != nil {
		rt.logger.Errorf("Failed to update target with id %d: %s", target.Id, err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
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

	ctx.AbortWithStatus(http.StatusOK)
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
