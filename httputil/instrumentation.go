package httputil

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mimir-news/mimir-go/id"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const metricsPath = "/metrics"

// Header keys
const (
	RequestIDHeader = "X-Request-ID"
	AcceptLanguage  = "Accept-Language"
)

// Default values
const (
	DefaltLocale = "sv"
)

// Prometheus metrics.
var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "The total number served requests",
		},
		[]string{"endpoint", "method", "status"},
	)
	requestsLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_latency_ms",
			Help: "Request latency in milliseconds",
		},
		[]string{"endpoint", "method", "status"},
	)
)

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// Metrics records metrics about a request.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == metricsPath {
			c.Next()
			return
		}
		stop := createTimer()
		endpoint := c.FullPath()
		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		latency := stop()
		requestsTotal.WithLabelValues(endpoint, method, status).Inc()
		requestsLatency.WithLabelValues(endpoint, method, status).Observe(latency)
	}
}

// RequestID annotates request with unique request id.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = id.New()
		}

		c.Set(RequestIDHeader, requestID)
		c.Header(RequestIDHeader, requestID)
		c.Next()
	}
}

// GetRequestID gets the request id from the gin context.
func GetRequestID(c *gin.Context) string {
	return c.GetString(RequestIDHeader)
}

// Locale extracts the user supplied Accept-Language headers and adds it to the request.
func Locale() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader(AcceptLanguage)
		if lang == "" {
			lang = DefaltLocale
		}

		c.Set(AcceptLanguage, lang)
		c.Next()
	}
}

// GetLocale gets the locale string from the gin context.
func GetLocale(c *gin.Context) string {
	return c.GetString(AcceptLanguage)
}

type calcDuration func() float64

func createTimer() calcDuration {
	start := time.Now()

	// Returns latency in milliseconds.
	return func() float64 {
		end := time.Now()
		return float64(end.Sub(start)) / 1e6
	}
}
