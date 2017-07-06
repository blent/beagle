package routes

import (
	"github.com/blent/beagle/src/core/notification"
	"github.com/blent/beagle/src/server/storage"
	"github.com/blent/beagle/src/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"path"
)

var (
	ErrEndpointsRouteInvalidEndpoint = errors.New("invalid endpoint")
)

type EndpointsRoute struct {
	baseUrl string
	logger  *zap.Logger
	storage *storage.Manager
}

func NewEndpointsRoute(baseUrl string, logger *zap.Logger, storage *storage.Manager) *EndpointsRoute {
	return &EndpointsRoute{baseUrl, logger, storage}
}

func (rt *EndpointsRoute) Use(routes gin.IRoutes) {
	singular := "endpoint"
	plural := "endpoints"

	// Get multiple endpoints
	routes.GET(path.Join("/", rt.baseUrl, plural), rt.findEndpoints)

	// Get single endpoint by id
	routes.GET(path.Join("/", rt.baseUrl, singular, ":id"), rt.getEndpoint)

	// Create new endpoint
	routes.POST(path.Join("/", rt.baseUrl, singular), rt.createEndpoint)

	// Update existing endpoint by id
	routes.PUT(path.Join("/", rt.baseUrl, singular), rt.updateEndpoint)

	// Delete existing endpoint by id
	routes.DELETE(path.Join("/", rt.baseUrl, singular, ":id"), rt.deleteEndpoint)

	// Delete existing endpoints by id
	routes.DELETE(path.Join("/", rt.baseUrl, plural), rt.deleteEndpoints)
}

func (rt *EndpointsRoute) findEndpoints(ctx *gin.Context) {
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

	name := ctx.Query("name")

	endpoints, quantity, err := rt.storage.FindEndpoints(storage.NewEndpointQuery(take, skip, name))

	if err != nil {
		rt.logger.Error("failed to find endpoints", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"items":    endpoints,
		"quantity": quantity,
	})
}

func (rt *EndpointsRoute) getEndpoint(ctx *gin.Context) {
	id, err := utils.StringToUint64(ctx.Params.ByName("id"))

	if err != nil {
		rt.logger.Error("Failed to parse endpoint id", zap.Error(err))
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	endpoint, err := rt.storage.GetEndpoint(id)

	if err != nil {
		rt.logger.Error(
			"Failed to retrieve endpoint",
			zap.Uint64("id", id),
			zap.Error(err),
		)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if endpoint == nil {
		ctx.AbortWithStatus(http.StatusOK)
		return
	}

	ctx.JSON(http.StatusOK, endpoint)
}

func (rt *EndpointsRoute) createEndpoint(ctx *gin.Context) {
	endpoint, ok := rt.deserializeEndpoint(ctx)

	if !ok {
		return
	}

	id, err := rt.storage.CreateEndpoint(endpoint)

	if err != nil {
		rt.logger.Error("Failed to create a new endpoint", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.String(http.StatusOK, "%d", id)
}

func (rt *EndpointsRoute) updateEndpoint(ctx *gin.Context) {
	endpoint, ok := rt.deserializeEndpoint(ctx)

	if !ok {
		return
	}

	if endpoint.Id == 0 {
		rt.logger.Error("Missed endpoint id")
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	err := rt.storage.UpdateEndpoint(endpoint)

	if err != nil {
		rt.logger.Error(
			"Failed to update endpoint",
			zap.Uint64("id", endpoint.Id),
			zap.Error(err),
		)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

func (rt *EndpointsRoute) deleteEndpoint(ctx *gin.Context) {
	id, err := utils.StringToUint64(ctx.Params.ByName("id"))

	if err != nil {
		rt.logger.Error("Failed to parse an endpoint id", zap.Error(err))
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	err = rt.storage.DeleteEndpoint(id)

	if err != nil {
		rt.logger.Error(
			"Failed to delete endpoint",
			zap.Uint64("id", id),
			zap.Error(err),
		)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

func (rt *EndpointsRoute) deleteEndpoints(ctx *gin.Context) {
	var ids []uint64

	err := ctx.BindJSON(&ids)

	if err != nil {
		rt.logger.Error("Failed to parse an array of endpoint ids", zap.Error(err))
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id(s)"))
		return
	}

	err = rt.storage.DeleteEndpoints(ids)

	if err != nil {
		rt.logger.Error(
			"Failed to delete endpoints",
			zap.Uint64s("ids", ids),
			zap.Error(err),
		)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

func (rt *EndpointsRoute) deserializeEndpoint(ctx *gin.Context) (*notification.Endpoint, bool) {
	var endpoint *notification.Endpoint

	err := ctx.BindJSON(&endpoint)

	if err != nil {
		rt.logger.Error("Failed to deserialize endpoint", zap.Error(err))
		ctx.AbortWithError(http.StatusBadRequest, ErrEndpointsRouteInvalidEndpoint)

		return nil, false
	}

	return endpoint, true
}
