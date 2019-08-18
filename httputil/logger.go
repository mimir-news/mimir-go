package httputil

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mimir-news/mimir-go/logger"
	"go.uber.org/zap"
)

var requestLog = logger.GetDefaultLogger("go24/requestLog")

// Logger request logging middleware.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		requestID := GetRequestID(c)
		requestLog.Info(fmt.Sprintf("Incomming request: %s %s", c.Request.Method, path), zap.String("requestId", requestID))

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		requestLog.Info(fmt.Sprintf("Outgoing request: %s %s", c.Request.Method, path),
			zap.Int("status", c.Writer.Status()),
			zap.String("requestId", requestID),
			zap.Duration("latency", latency))
	}
}

// Trace logs a message along with request id.
func Trace(c *gin.Context, message string) {
	requestLog.Info(message, zap.String("requestId", GetRequestID(c)))
}
