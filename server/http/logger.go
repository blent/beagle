package http

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Start timer
		start := time.Now()
		path := ctx.Request.URL.Path
		method := ctx.Request.Method

		logger.Info(
			"Started",
			zap.String("path", path),
			zap.String("method", method),
		)

		ctx.Next()

		logger.Info(
			"Completed",
			zap.String("path", path),
			zap.String("method", method),
			zap.Int("status", ctx.Writer.Status()),
			zap.String("status-text", http.StatusText(ctx.Writer.Status())),
			zap.Duration("time", time.Since(start)),
		)
	}
}
