package http

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
	"fmt"
)

func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Start timer
		start := time.Now()
		path := ctx.Request.URL.Path
		method := ctx.Request.Method

		ctx.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		logger.Info("Incoming HTTP request",
			zap.String("time", end.Format("2006/01/02 15:04:05")),
			zap.Int("status", ctx.Writer.Status()),
			zap.String("latency", fmt.Sprintf("%13v", latency)),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("clientIP", ctx.ClientIP()),
		)
	}
}
