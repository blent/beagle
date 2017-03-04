package routes

import (
	"fmt"
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/notification"
	"github.com/blent/beagle/src/server/http/routes/dto"
	"github.com/blent/beagle/src/server/storage"
	"github.com/blent/beagle/src/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"path"
)

var (
	ErrEndpointRouteInvalidEndpoint = errors.New("invalid endpoint")
)

type EndpointRoute struct {
	baseUrl string
	logger  *logging.Logger
	storage *storage.Manager
}

func NewEndpointRoute(baseUrl string, logger *logging.Logger, storage *storage.Manager) *EndpointRoute {
	return &EndpointRoute{baseUrl, logger, storage}
}

func (rt *EndpointRoute) Use(routes gin.IRoutes) {
	route := "endpoint"

	// Get multiple endpoints
	routes.GET(path.Join("/", rt.baseUrl, route), rt.findEndpoints)

	// Get single endpoint by id
	routes.GET(path.Join("/", rt.baseUrl, route, ":id"), rt.getEndpoint)

	// Create new endpoint
	routes.POST(path.Join("/", rt.baseUrl, route), rt.createEndpoint)

	// Update existing endpoint by id
	routes.PUT(path.Join("/", rt.baseUrl, route), rt.updateEndpoint)

	// Delete existing endpoint by id
	routes.DELETE(path.Join("/", rt.baseUrl, route, ":id"), rt.deleteEndpoint)
}

func (rt *EndpointRoute) findEndpoints(ctx *gin.Context) {
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

	endpoints, err := rt.storage.FindEndpoints(storage.NewEndpointQuery(take, skip))

	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	targetsDto := make([]*dto.Endpoint, 0, len(endpoints))

	for _, target := range endpoints {
		endpointDto, ok := rt.serializeEndpoint(ctx, target)

		if !ok {
			return
		}

		targetsDto = append(targetsDto, endpointDto)
	}

	ctx.JSON(http.StatusOK, targetsDto)
}

func (rt *EndpointRoute) getEndpoint(ctx *gin.Context) {
	id, err := utils.StringToUint64(ctx.Params.ByName("id"))

	if err != nil {
		rt.logger.Error(fmt.Sprintf("Failed to parse endpoint id: %s", err.Error()))
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	endpoint, err := rt.storage.GetEndpoint(id)

	if err != nil {
		rt.logger.Error(fmt.Sprintf("Failed to retrieve endpoint %d: %s", id, err.Error()))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if endpoint == nil {
		ctx.AbortWithStatus(http.StatusOK)
		return
	}

	endpointDto, ok := rt.serializeEndpoint(ctx, endpoint)

	if !ok {
		return
	}

	ctx.JSON(http.StatusOK, endpointDto)
}

func (rt *EndpointRoute) createEndpoint(ctx *gin.Context) {
	endpoint, ok := rt.deserializeEndpoint(ctx)

	if !ok {
		return
	}

	id, err := rt.storage.CreateEndpoint(endpoint)

	if err != nil {
		rt.logger.Errorf("Failed to create a new endpoint: %s", err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.String(http.StatusOK, "%d", id)
}

func (rt *EndpointRoute) updateEndpoint(ctx *gin.Context) {
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
		rt.logger.Errorf("Failed to update endpoint with id %d: %s", endpoint.Id, err.Error())
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

func (rt *EndpointRoute) deleteEndpoint(ctx *gin.Context) {
	id, err := utils.StringToUint64(ctx.Params.ByName("id"))

	if err != nil {
		rt.logger.Error(fmt.Sprintf("Failed to parse endpoint id: %s", err.Error()))
		ctx.AbortWithError(http.StatusBadRequest, errors.New("missed id"))
		return
	}

	err = rt.storage.DeleteEndpoint(id)

	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

func (rt *EndpointRoute) serializeEndpoint(ctx *gin.Context, endpoint *notification.Endpoint) (*dto.Endpoint, bool) {
	endpointDto, err := dto.FromEndpoint(endpoint)

	if err != nil {
		rt.logger.Errorf("Failed to serialize endpoint: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrEndpointRouteInvalidEndpoint)

		return nil, false
	}

	return endpointDto, true
}

func (rt *EndpointRoute) deserializeEndpoint(ctx *gin.Context) (*notification.Endpoint, bool) {
	var endpointDto *dto.Endpoint

	err := ctx.BindJSON(&endpointDto)

	if err != nil {
		rt.logger.Errorf("Failed to deserialize endpoint: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrEndpointRouteInvalidEndpoint)

		return nil, false
	}

	endpoint, err := dto.ToEndpoint(endpointDto)

	if err != nil {
		rt.logger.Errorf("Failed to deserialize endpoint: %s", err.Error())
		ctx.AbortWithError(http.StatusBadRequest, ErrEndpointRouteInvalidEndpoint)

		return nil, false
	}

	return endpoint, true
}
