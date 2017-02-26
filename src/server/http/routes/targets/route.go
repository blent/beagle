package routes

import (
	"github.com/blent/beagle/src/core/logging"
	"github.com/blent/beagle/src/core/tracking"
	"github.com/blent/beagle/src/server/storage"
	"github.com/gin-gonic/gin"
	"net/http"
	"path"
	"strconv"
)

type TargetsRoute struct {
	baseEndpoint string
	logger       *logging.Logger
	repo         storage.TargetRepository
}

func NewTargetsRoute(baseEndpoint string, logger *logging.Logger, repo storage.TargetRepository) *TargetsRoute {
	return &TargetsRoute{baseEndpoint, logger, repo}
}

func (rt *TargetsRoute) Use(routes gin.IRoutes) {
	route := "targets"

	routes.GET(path.Join("/", rt.baseEndpoint, route, ":id"), rt.getTarget)
	routes.GET(path.Join("/", rt.baseEndpoint, route), rt.findTargets)
	routes.POST(path.Join("/", rt.baseEndpoint, route), rt.createTarget)
	routes.PUT(path.Join("/", rt.baseEndpoint, route, ":id"), rt.updateTarget)
	routes.DELETE(path.Join("/", rt.baseEndpoint, route, ":id"), rt.deleteTarget)
}

func (rt *TargetsRoute) getTarget(ctx *gin.Context) {
	id, ok := rt.parseId(ctx)

	if !ok {
		return
	}

	target, err := rt.repo.GetById(id)

	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, target)
}

func (rt *TargetsRoute) findTargets(ctx *gin.Context) {
	query := &storage.TargetQuery{
		Take:   10,
		Skip:   0,
		Status: storage.STATUS_ALL,
	}

	targets, err := rt.repo.Query(query)

	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, targets)
}

func (rt *TargetsRoute) createTarget(ctx *gin.Context) {
	target, ok := rt.parseTarget(ctx)

	if !ok {
		return
	}

	if err := rt.repo.Update(target); err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.String(http.StatusOK, "%n", target.ID)
}

func (rt *TargetsRoute) updateTarget(ctx *gin.Context) {
	target, ok := rt.parseTarget(ctx)

	if !ok {
		return
	}

	if err := rt.repo.Update(target); err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusNoContent)
}

func (rt *TargetsRoute) deleteTarget(ctx *gin.Context) {
	id, ok := rt.parseId(ctx)

	if !ok {
		return
	}

	err := rt.repo.Delete(id)

	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.AbortWithStatus(http.StatusNoContent)
}

func (rt *TargetsRoute) parseId(ctx *gin.Context) (uint64, bool) {
	idStr := ctx.Params.ByName("id")

	if idStr == "" {
		ctx.AbortWithError(http.StatusBadRequest, ErrMissedId)
		return 0, false
	}

	id, err := strconv.ParseUint(idStr, 10, 64)

	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, ErrInvalidId)
		return 0, false
	}

	return id, true
}

func (rt *TargetsRoute) parseTarget(ctx *gin.Context) (*tracking.Target, bool) {
	var target *tracking.Target

	if err := ctx.BindJSON(&target); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, ErrInvalidTarget)
		return nil, false
	}

	return target, true
}
