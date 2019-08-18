package httputil

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthFunc health check function signature.
type HealthFunc func() error

// NewRouter creates a default router.
func NewRouter(healthCheck HealthFunc) *gin.Engine {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		Metrics(),
		RequestID(),
		Locale(),
		Logger(),
		HandleErrors())

	r.GET("/health", checkHealth(healthCheck))
	r.GET(metricsPath, prometheusHandler())
	return r
}

// SendOK sends an ok status and message to the client.
func SendOK(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

// ParseQueryValue parses a query value from request.
func ParseQueryValue(c *gin.Context, key string) (string, error) {
	value, ok := c.GetQuery(key)
	if !ok {
		errorMessage := fmt.Sprintf("No value found for param: %s", key)
		return "", BadRequest(errorMessage)
	}
	return value, nil
}

// ParseQueryValues parses query values from a request.
func ParseQueryValues(c *gin.Context, key string) ([]string, error) {
	values, ok := c.GetQueryArray(key)
	if !ok {
		errorMessage := fmt.Sprintf("No value found for param: %s", key)
		return nil, BadRequest(errorMessage)
	}
	return values, nil
}

func checkHealth(check HealthFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := check()
		if err == nil {
			SendOK(c)
			return
		}

		c.Error(err)
	}
}
