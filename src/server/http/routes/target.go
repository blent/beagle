package routes

import (
	"fmt"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/notification"
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
	baseUrl string
	logger  *logging.Logger
	storage *storage.Manager
}

func NewTargetRoute(baseUrl string, logger *logging.Logger, storage *storage.Manager) *TargetRoute {
	return &TargetRoute{
		baseUrl,
		logger,
		storage,
	}
}

func (rt *TargetRoute) Use(routes gin.IRoutes) {
	route := "target"

	// Get multiple targets
	routes.GET(path.Join("/", rt.baseUrl, route), rt.findTargets)

	// Get single target by id
	routes.GET(path.Join("/", rt.baseUrl, route, ":id"), rt.getTarget)

	// Create new target
	routes.POST(path.Join("/", rt.baseUrl, route), rt.createTarget)

	// Update existing target by id
	routes.PUT(path.Join("/", rt.baseUrl, route), rt.updateTarget)

	// Delete existing target by id
	routes.DELETE(path.Join("/", rt.baseUrl, route, ":id"), rt.deleteTarget)
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

	targets, err := rt.storage.FindTargets(storage.NewTargetQuery(take, skip, storage.TARGET_STATUS_ANY))

	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	targetsDto := make([]*dto.Target, 0, len(targets))

	for _, target := range targets {
		targetDto, ok := rt.serializeTarget(ctx, target, nil)

		if !ok {
			return
		}

		targetsDto = append(targetsDto, targetDto)
	}

	ctx.JSON(http.StatusOK, targetsDto)
}

func (rt *TargetRoute) getTarget(ctx *gin.Context) {
	id, err := utils.StringToUint64(ctx.Params.ByName("id"))

	if err != nil {
		rt.logger.Error(fmt.Sprintf("Failed to parse target id: %s", err.Error()))
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	target, subscribers, err := rt.storage.GetTargetWithSubscribers(id)

	if err != nil {
		rt.logger.Error(fmt.Sprintf("Failed to retrieve target %d: %s", id, err.Error()))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if target == nil {
		ctx.AbortWithStatus(http.StatusOK)
		return
	}

	targetDto, ok := rt.serializeTarget(ctx, target, subscribers)

	if !ok {
		return
	}

	ctx.JSON(http.StatusOK, targetDto)
}

func (rt *TargetRoute) createTarget(ctx *gin.Context) {
	target, subscribers, ok := rt.deserializeTarget(ctx)

	if !ok {
		return
	}

	id, err := rt.storage.CreateTarget(target, subscribers)

	if err != nil {
		rt.logger.Errorf("Failed to create new target: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.String(http.StatusOK, "%d", id)
}

func (rt *TargetRoute) updateTarget(ctx *gin.Context) {
	target, subscribers, ok := rt.deserializeTarget(ctx)

	if !ok {
		return
	}

	if target.Id < 0 || target.Id == 0 {
		rt.logger.Error("Missed target id")
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	err := rt.storage.UpdateTarget(target, subscribers)

	if err != nil {
		rt.logger.Errorf("Failed to update target with id %d: %s", target.Id, err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

func (rt *TargetRoute) deleteTarget(ctx *gin.Context) {
	id, err := utils.StringToUint64(ctx.Params.ByName("id"))

	if err != nil {
		rt.logger.Error(fmt.Sprintf("Failed to parse target id: %s", err.Error()))
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	err = rt.storage.DeleteTarget(id)

	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

func (rt *TargetRoute) serializeTarget(ctx *gin.Context, target *tracking.Target, subscribers []*notification.Subscriber) (*dto.Target, bool) {
	targetDto, err := dto.FromTarget(target)

	if err != nil {
		rt.logger.Errorf("Failed to serialize target: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrTargetRouteInvalidTarget)

		return nil, false
	}

	if subscribers != nil {
		targetDto.Subscribers = make([]*dto.Subscriber, 0, len(subscribers))

		for _, subscriber := range subscribers {
			subDto, failure := dto.FromSubscriber(subscriber)

			if failure != nil {
				err = failure
				break
			}

			targetDto.Subscribers = append(targetDto.Subscribers, subDto)
		}
	}

	if err != nil {
		rt.logger.Errorf("Failed to serialize subscribers: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrTargetRouteInvalidTarget)

		return nil, false
	}

	return targetDto, true
}

func (rt *TargetRoute) deserializeTarget(ctx *gin.Context) (*tracking.Target, []*notification.Subscriber, bool) {
	var targetDto *dto.Target

	err := ctx.BindJSON(&targetDto)

	if err != nil {
		rt.logger.Errorf("Failed to deserialize target: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrTargetRouteInvalidTarget)

		return nil, nil, false
	}

	target, err := dto.ToTarget(targetDto)

	if err != nil {
		rt.logger.Errorf("Failed to deserialize target: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrTargetRouteInvalidTarget)

		return nil, nil, false
	}

	var subscribers []*notification.Subscriber

	if targetDto.Subscribers != nil && len(targetDto.Subscribers) > 0 {
		subscribers = make([]*notification.Subscriber, 0, len(targetDto.Subscribers))

		for _, subDto := range targetDto.Subscribers {
			subscriber, failure := dto.ToSubscriber(subDto)

			if failure != nil {
				err = failure
				break
			}

			subscribers = append(subscribers, subscriber)
		}
	}

	if err != nil {
		rt.logger.Errorf("Failed to deserialize subscriber: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrTargetRouteInvalidTarget)

		return nil, nil, false
	}

	return target, subscribers, true
}
