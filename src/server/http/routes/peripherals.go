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
	ErrPeripheralsRouteInvalidModel = errors.New("invalid peripheral")
)

type PeripheralsRoute struct {
	baseUrl string
	logger  *logging.Logger
	storage *storage.Manager
}

func NewPeripheralsRoute(baseUrl string, logger *logging.Logger, storage *storage.Manager) *PeripheralsRoute {
	return &PeripheralsRoute{
		baseUrl,
		logger,
		storage,
	}
}

func (rt *PeripheralsRoute) Use(routes gin.IRoutes) {
	singular := "peripheral"
	plural := "peripherals"

	// Get multiple targets
	routes.GET(path.Join("/", rt.baseUrl, plural), rt.findPeripherals)

	// Get single target by id
	routes.GET(path.Join("/", rt.baseUrl, singular, ":id"), rt.getPeripheral)

	// Create new target
	routes.POST(path.Join("/", rt.baseUrl, singular), rt.createPeripheral)

	// Update existing target by id
	routes.PUT(path.Join("/", rt.baseUrl, singular), rt.updatePeripheral)

	// Delete existing target by id
	routes.DELETE(path.Join("/", rt.baseUrl, singular, ":id"), rt.deletePeripheral)
}

func (rt *PeripheralsRoute) findPeripherals(ctx *gin.Context) {
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

	targets, quantity, err := rt.storage.FindPeripherals(storage.NewTargetQuery(take, skip, storage.PERIPHERAL_STATUS_ANY))

	if err != nil {
		rt.logger.Errorf("failed to find peripherals: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	peripheralsDto := make([]interface{}, 0, len(targets))

	for _, target := range targets {
		targetDto, ok := rt.serializePeripheral(ctx, target, nil)

		if !ok {
			return
		}

		peripheralsDto = append(peripheralsDto, targetDto)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items":    peripheralsDto,
		"quantity": quantity,
	})
}

func (rt *PeripheralsRoute) getPeripheral(ctx *gin.Context) {
	id, err := utils.StringToUint64(ctx.Params.ByName("id"))

	if err != nil {
		rt.logger.Error(fmt.Sprintf("Failed to parse peripheral id: %s", err.Error()))
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	target, subscribers, err := rt.storage.GetPeripheralWithSubscribers(id)

	if err != nil {
		rt.logger.Error(fmt.Sprintf("Failed to retrieve peripheral %d: %s", id, err.Error()))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if target == nil {
		ctx.AbortWithStatus(http.StatusOK)
		return
	}

	targetDto, ok := rt.serializePeripheral(ctx, target, subscribers)

	if !ok {
		return
	}

	ctx.JSON(http.StatusOK, targetDto)
}

func (rt *PeripheralsRoute) createPeripheral(ctx *gin.Context) {
	target, subscribers, ok := rt.deserializePeripheral(ctx)

	if !ok {
		return
	}

	id, err := rt.storage.CreatePeripheral(target, subscribers)

	if err != nil {
		rt.logger.Errorf("Failed to create new peripheral: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.String(http.StatusOK, "%d", id)
}

func (rt *PeripheralsRoute) updatePeripheral(ctx *gin.Context) {
	target, subscribers, ok := rt.deserializePeripheral(ctx)

	if !ok {
		return
	}

	if target.Id < 0 || target.Id == 0 {
		rt.logger.Error("Missed peripheral id")
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	err := rt.storage.UpdatePeripheral(target, subscribers)

	if err != nil {
		rt.logger.Errorf("Failed to update peripheral with id %d: %s", target.Id, err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

func (rt *PeripheralsRoute) deletePeripheral(ctx *gin.Context) {
	id, err := utils.StringToUint64(ctx.Params.ByName("id"))

	if err != nil {
		rt.logger.Error(fmt.Sprintf("Failed to parse peripheral id: %s", err.Error()))
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	err = rt.storage.DeletePeripheral(id)

	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

func (rt *PeripheralsRoute) serializePeripheral(ctx *gin.Context, target *tracking.Peripheral, subscribers []*notification.Subscriber) (dto.Peripheral, bool) {
	targetDto, err := dto.FromPeripheral(target)

	if err != nil {
		rt.logger.Errorf("Failed to serialize peripheral: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrPeripheralsRouteInvalidModel)

		return nil, false
	}

	if subscribers != nil {
		dtoSubscribers := make([]*dto.Subscriber, 0, len(subscribers))

		for _, subscriber := range subscribers {
			subDto, failure := dto.FromSubscriber(subscriber)

			if failure != nil {
				err = failure
				break
			}

			dtoSubscribers = append(dtoSubscribers, subDto)
		}

		targetDto.SetSubscribers(dtoSubscribers)
	}

	if err != nil {
		rt.logger.Errorf("Failed to serialize subscribers: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrPeripheralsRouteInvalidModel)

		return nil, false
	}

	return targetDto, true
}

func (rt *PeripheralsRoute) deserializePeripheral(ctx *gin.Context) (*tracking.Peripheral, []*notification.Subscriber, bool) {
	var err error
	var input map[string]interface{}

	err = ctx.BindJSON(&input)

	if err != nil {
		rt.logger.Errorf("Failed to deserialize peripheral: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrPeripheralsRouteInvalidModel)

		return nil, nil, false
	}

	peripheralDto, err := dto.FromPeripheralMap(input)

	if err != nil {
		rt.logger.Errorf("Failed to deserialize peripheral: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrPeripheralsRouteInvalidModel)

		return nil, nil, false
	}

	peripheral, err := dto.ToPeripheral(peripheralDto)

	if err != nil {
		rt.logger.Errorf("Failed to deserialize peripheral: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrPeripheralsRouteInvalidModel)

		return nil, nil, false
	}

	if peripheral == nil {
		rt.logger.Error("Missed peripheral")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return nil, nil, false
	}

	var subscribers []*notification.Subscriber

	if peripheralDto.GetSubscribers() != nil && len(peripheralDto.GetSubscribers()) > 0 {
		subscribers = make([]*notification.Subscriber, 0, len(peripheralDto.GetSubscribers()))

		for _, subDto := range peripheralDto.GetSubscribers() {
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
		ctx.AbortWithError(http.StatusBadRequest, ErrPeripheralsRouteInvalidModel)

		return nil, nil, false
	}

	return peripheral, subscribers, true
}
