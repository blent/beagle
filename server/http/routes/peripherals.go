package routes

import (
	"github.com/blent/beagle/pkg/discovery/peripherals"
	"github.com/blent/beagle/pkg/notification"
	"github.com/blent/beagle/pkg/tracking"
	"github.com/blent/beagle/server/storage"
	"github.com/blent/beagle/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"path"
	"strings"
)

var (
	ErrPeripheralsRouteInvalidModel = errors.New("invalid peripheral")
)

type (
	// We make one big generic DTO for all types of Peripherals
	// Just to make deserialization more simple and fast
	Dto struct {
		Id          uint64                     `json:"id"`
		Kind        string                     `json:"kind" binding:"required"`
		Name        string                     `json:"name" binding:"required"`
		Enabled     bool                       `json:"enabled"`
		Uuid        string                     `json:"uuid, omitempty"`
		Major       uint16                     `json:"major, omitempty"`
		Minor       uint16                     `json:"minor, omitempty"`
		Subscribers []*notification.Subscriber `json:"subscribers"`
	}

	PeripheralsRoute struct {
		baseUrl string
		logger  *zap.Logger
		storage *storage.Manager
	}
)

func NewPeripheralsRoute(baseUrl string, logger *zap.Logger, storage *storage.Manager) *PeripheralsRoute {
	return &PeripheralsRoute{
		baseUrl,
		logger,
		storage,
	}
}

func (rt *PeripheralsRoute) Use(routes gin.IRoutes) {
	singular := "peripheral"
	plural := "peripherals"

	// Get multiple peripherals
	routes.GET(path.Join("/", rt.baseUrl, plural), rt.findPeripherals)

	// Get single peripheral by id
	routes.GET(path.Join("/", rt.baseUrl, singular, ":id"), rt.getPeripheral)

	// Create new peripheral
	routes.POST(path.Join("/", rt.baseUrl, singular), rt.createPeripheral)

	// Update existing peripheral by id
	routes.PUT(path.Join("/", rt.baseUrl, singular), rt.updatePeripheral)

	// Delete existing peripheral by id
	routes.DELETE(path.Join("/", rt.baseUrl, singular, ":id"), rt.deletePeripheral)

	// Delete multiple peripherals by id
	routes.DELETE(path.Join("/", rt.baseUrl, plural), rt.deletePeripherals)
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
		rt.logger.Error("failed to find peripherals", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items":    targets,
		"quantity": quantity,
	})
}

func (rt *PeripheralsRoute) getPeripheral(ctx *gin.Context) {
	id, err := utils.StringToUint64(ctx.Params.ByName("id"))

	if err != nil {
		rt.logger.Error("Failed to parse peripheral id", zap.Error(err))
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	target, subscribers, err := rt.storage.GetPeripheralWithSubscribers(id)

	if err != nil {
		rt.logger.Error(
			"Failed to retrieve peripheral",
			zap.Uint64("id", id),
			zap.Error(err),
		)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if target == nil {
		ctx.AbortWithStatus(http.StatusOK)
		return
	}

	dto, err := rt.serializePeripheral(target, subscribers)

	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, ErrPeripheralsRouteInvalidModel)
		return
	}

	ctx.JSON(http.StatusOK, dto)
}

func (rt *PeripheralsRoute) createPeripheral(ctx *gin.Context) {
	target, subscribers, err := rt.deserializePeripheral(ctx)

	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id, err := rt.storage.CreatePeripheral(target, subscribers)

	if err != nil {
		rt.logger.Error("Failed to create new peripheral", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.String(http.StatusOK, "%d", id)
}

func (rt *PeripheralsRoute) updatePeripheral(ctx *gin.Context) {
	target, subscribers, err := rt.deserializePeripheral(ctx)

	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if target.Id < 0 || target.Id == 0 {
		rt.logger.Error("Missed peripheral id")
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	err = rt.storage.UpdatePeripheral(target, subscribers)

	if err != nil {
		rt.logger.Error(
			"Failed to update peripheral",
			zap.Uint64("id", target.Id),
			zap.String("error", err.Error()),
		)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

func (rt *PeripheralsRoute) deletePeripheral(ctx *gin.Context) {
	id, err := utils.StringToUint64(ctx.Params.ByName("id"))

	if err != nil {
		rt.logger.Error("Failed to parse peripheral id", zap.Error(err))
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	err = rt.storage.DeletePeripheral(id)

	if err != nil {
		rt.logger.Error(
			"Failed to delete peripheral",
			zap.Uint64("id", id),
			zap.Error(err),
		)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

func (rt *PeripheralsRoute) deletePeripherals(ctx *gin.Context) {
	var ids []uint64

	err := ctx.BindJSON(&ids)

	if err != nil {
		rt.logger.Error("Failed to parse an array of peripheral ids", zap.Error(err))
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id(s)"))
		return
	}

	err = rt.storage.DeletePeripherals(ids)

	if err != nil {
		rt.logger.Error(
			"Failed to delete peripherals",
			zap.Uint64s("ids", ids),
			zap.Error(err),
		)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

func (rt *PeripheralsRoute) serializePeripheral(target *tracking.Peripheral, subscribers []*notification.Subscriber) (*Dto, error) {
	var err error

	dto := &Dto{
		Id:          target.Id,
		Kind:        target.Kind,
		Name:        target.Name,
		Enabled:     target.Enabled,
		Subscribers: subscribers,
	}

	switch target.Kind {
	case peripherals.PERIPHERAL_IBEACON:
		uuid, major, minor, err := peripherals.ParseIBeaconUniqueKey(target.Key)

		if err != nil {
			return nil, err
		}

		dto.Uuid = uuid
		dto.Major = major
		dto.Minor = minor
	default:
		err = errors.Errorf("unsupported peripheral kind: '%s'", target.Kind)
	}

	if err != nil {
		rt.logger.Error("Failed to serialize peripheral", zap.Error(err))

		return nil, err
	}

	return dto, nil
}

func (rt *PeripheralsRoute) deserializePeripheral(ctx *gin.Context) (*tracking.Peripheral, []*notification.Subscriber, error) {
	var err error
	var dto Dto

	err = ctx.BindJSON(&dto)

	if err != nil {
		rt.logger.Error("Failed to deserialize peripheral", zap.Error(err))

		return nil, nil, err
	}

	var key string

	switch dto.Kind {
	case peripherals.PERIPHERAL_IBEACON:
		dto.Uuid = strings.TrimSpace(dto.Uuid)

		if len(dto.Uuid) != 32 {
			err = errors.Errorf("invalid uuid length: %d", len(dto.Uuid))
			break
		}

		if dto.Major == 0 {
			err = errors.Errorf("invalid major number: %d", dto.Major)
			break
		}

		if dto.Minor == 0 {
			err = errors.Errorf("invalid minor number: %d", dto.Minor)
			break
		}

		key = peripherals.CreateIBeaconUniqueKey(dto.Uuid, dto.Major, dto.Minor)
	default:
		err = errors.Errorf("unsupported peripheral kind: '%s'", dto.Kind)
	}

	if err != nil {
		return nil, nil, err
	}

	peripheral := &tracking.Peripheral{
		Id:      dto.Id,
		Key:     key,
		Name:    dto.Name,
		Kind:    dto.Kind,
		Enabled: dto.Enabled,
	}

	return peripheral, dto.Subscribers, nil
}
