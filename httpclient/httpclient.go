package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mimir-news/mimir-go/context"
	"github.com/mimir-news/mimir-go/httputil"
	"github.com/mimir-news/mimir-go/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var log = logger.GetDefaultLogger("mimir-go/httpclient").Sugar()

// Prometheus metrics.
var (
	rpcsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rpc_requests_total",
			Help: "The total number of remote procedure calls",
		},
		[]string{"endpoint", "method", "status"},
	)
	rpcLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "rpc_request_latency_ms",
			Help: "Remote procedure call duration in milliseconds",
		},
		[]string{"endpoint", "method", "status"},
	)
)

// Client interface for http client.
type Client interface {
	Get(ctx *context.Context, path string) (*http.Response, error)
	Post(ctx *context.Context, path string, body interface{}) (*http.Response, error)
	Put(ctx *context.Context, path string, body interface{}) (*http.Response, error)
	Delete(ctx *context.Context, path string) (*http.Response, error)
	Request(ctx *context.Context, path, method string, body interface{}) (*http.Response, error)
}

type remoteError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type client struct {
	baseURL          string
	name             string
	httpClient       *http.Client
	warningThreshold time.Duration
}

// New creates a httpclient.
func New(name, baseURL string, threshold time.Duration) Client {
	return &client{
		name:             name,
		baseURL:          baseURL,
		httpClient:       http.DefaultClient,
		warningThreshold: threshold,
	}
}

func (c *client) Get(ctx *context.Context, path string) (*http.Response, error) {
	log.Debugw("client.Get", "client", c.name, "path", stripQueryParameters(path), "ctx", ctx)
	return c.Request(ctx, path, http.MethodGet, nil)
}

func (c *client) Post(ctx *context.Context, path string, body interface{}) (*http.Response, error) {
	log.Debugw("client.Post", "client", c.name, "path", stripQueryParameters(path), "ctx", ctx)
	return c.Request(ctx, path, http.MethodPost, body)
}

func (c *client) Put(ctx *context.Context, path string, body interface{}) (*http.Response, error) {
	log.Debugw("client.Put", "client", c.name, "path", stripQueryParameters(path), "ctx", ctx)
	return c.Request(ctx, path, http.MethodPut, body)
}

func (c *client) Delete(ctx *context.Context, path string) (*http.Response, error) {
	log.Debugw("client.Delete", "client", c.name, "path", stripQueryParameters(path), "ctx", ctx)
	return c.Request(ctx, path, http.MethodDelete, nil)
}

func (c *client) Request(ctx *context.Context, path, method string, body interface{}) (*http.Response, error) {
	startTime := time.Now()
	timer := createTimer(startTime)
	defer c.logRequestLatency(ctx, startTime, path)
	req, err := c.createRequest(ctx, path, method, body)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil || (res != nil && res.StatusCode >= 300) {
		c.recordMetricsOnError(timer, path, method, res)
		return nil, c.wrapError(ctx, res, err)
	}

	c.recordMetrics(timer, path, method, res.StatusCode)
	return res, nil
}

func (c *client) createRequest(ctx *context.Context, path, method string, body interface{}) (*http.Request, error) {
	fullURL := c.baseURL + path
	bodyReader, err := createBody(body)
	if err != nil {
		c.logError(ctx, "Failed to create request body", method, path, err)
		return nil, err
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		c.logError(ctx, "Failed to create request body", method, path, err)
		return nil, err
	}

	if ctx.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+ctx.AuthToken)
	}

	req.Header.Set("X-ClientID", ctx.ClientID)
	req.Header.Set(httputil.RequestIDHeader, ctx.ID)
	req.Header.Set(httputil.AcceptLanguage, ctx.Language)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *client) wrapError(ctx *context.Context, res *http.Response, err error) error {
	if res == nil {
		message := fmt.Sprintf("Downstream request failed with no response. requestId=[%s] err=[%s]", ctx.ID, err)
		return httputil.BadGateway(message)
	}

	var remoteErr remoteError
	parseErr := json.NewDecoder(res.Body).Decode(&remoteErr)
	if res == nil {
		log.Warnw("Failed to parse remoteError", "client", c.name, "requestId", ctx.ID, "error", parseErr)
		message := fmt.Sprintf("Downstream request failed could not parse error. requestId=[%s] code=[%d] err=[%s]", ctx.ID, res.StatusCode, err)
		return httputil.BadGateway(message)
	}

	message := fmt.Sprintf("Downstream failed. requestId=[%s] status=[%d] code=[%d] message=[%s]", ctx.ID, res.StatusCode, remoteErr.Code, remoteErr.Message)
	return httputil.BadGateway(message)
}

func (c *client) logError(ctx *context.Context, message, method, path string, err error) {
	log.Errorw(message, "client", c.name, "method", method, "url", stripQueryParameters(path), "ctx", ctx, "error", err)
}

func (c *client) logRequestLatency(ctx *context.Context, startTime time.Time, path string) {
	duration := time.Now().Sub(startTime)
	if duration < c.warningThreshold {
		return
	}

	latency := fmt.Sprintf("%.2f ms", toMilliseconds(duration))
	threshold := fmt.Sprintf("%.2f ms", toMilliseconds(c.warningThreshold))
	log.Warnw("Unusually high latency in service call",
		"path", stripQueryParameters(path),
		"requestId", ctx.ID,
		"clientId", ctx.ClientID,
		"latency", latency,
		"warningThreshold", threshold)
}

func (c *client) recordMetricsOnError(timer calcDuration, path, method string, res *http.Response) {
	statusCode := http.StatusServiceUnavailable
	if res != nil {
		statusCode = res.StatusCode
	}

	c.recordMetrics(timer, path, method, statusCode)
}

func (c *client) recordMetrics(stopTimer calcDuration, path, method string, statusCode int) {
	latency := stopTimer()
	endpoint := stripQueryAndUUIDs(c.baseURL + path)
	status := strconv.Itoa(statusCode)

	rpcsTotal.WithLabelValues(endpoint, method, status).Inc()
	rpcLatency.WithLabelValues(endpoint, method, status).Observe(latency)
}

func createBody(body interface{}) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}

	var bodyReader io.Reader
	if body != nil {
		bytesBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(bytesBody)
	}

	return bodyReader, nil
}

type calcDuration func() float64

func createTimer(start time.Time) calcDuration {
	// Returns latency in milliseconds.
	return func() float64 {
		end := time.Now()
		return toMilliseconds(end.Sub(start))
	}
}

func toMilliseconds(d time.Duration) float64 {
	return float64(d) / 1e6
}

var uuidRegexp = regexp.MustCompile(`[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}`)

func stripQueryAndUUIDs(url string) string {
	withoutQuery := strings.Split(url, "?")[0]
	return uuidRegexp.ReplaceAllString(withoutQuery, ":id")
}

func stripQueryParameters(url string) string {
	split := strings.Split(url, "?")
	domainAndPath := split[0]
	if len(split) != 2 {
		return domainAndPath
	}

	query := strings.Split(split[1], "&")
	strippedQuery := make([]string, 0, len(query))
	for _, part := range query {
		paramAndValue := strings.Split(part, "=")
		param := paramAndValue[0]
		stripped := ":value"
		if len(paramAndValue) == 2 && len(strings.Split(paramAndValue[1], ",")) > 1 {
			stripped = ":values"
		}
		strippedQuery = append(strippedQuery, fmt.Sprintf("%s=%s", param, stripped))
	}

	return fmt.Sprintf("%s?%s", domainAndPath, strings.Join(strippedQuery, "&"))
}
