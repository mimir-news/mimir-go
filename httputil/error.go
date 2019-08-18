package httputil

import (
	"fmt"
	"net/http"

	"github.com/mimir-news/mimir-go/id"
	"github.com/mimir-news/mimir-go/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var errLog = logger.GetDefaultLogger("mimir-go/errorLog")

// HandleErrors wrapper function to deal with encountered errors
// during request handling.
func HandleErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		err := getFirstError(c)
		if err == nil {
			return
		}

		httpError := NewErrorResponse(c, err)
		logError(c, httpError)
		sendError(c, httpError)
	}
}

// Error implements the error interface with a message, id and http status code.
type Error struct {
	ID         string `json:"id,omitempty"`
	Message    string `json:"message,omitempty"`
	StatusCode int    `json:"status,omitempty"`
}

func (err *Error) Error() string {
	return fmt.Sprintf("Error(id=%s, statusCode=%d message=[%s])", err.ID, err.StatusCode, err.Message)
}

// NewError creates a new error.
func NewError(message string, status int) *Error {
	errMsg := message
	if message == "" {
		errMsg = http.StatusText(status)
	}

	return &Error{
		ID:         id.New(),
		Message:    errMsg,
		StatusCode: status,
	}
}

// BadRequest creates a new bad request (400) error.
func BadRequest(message string) *Error {
	return NewError(message, http.StatusBadRequest)
}

// Unauthorized creates a new unauthorized (401) error.
func Unauthorized(message string) *Error {
	return NewError(message, http.StatusUnauthorized)
}

// Forbidden creates a new forbidden (402) error.
func Forbidden(message string) *Error {
	return NewError(message, http.StatusForbidden)
}

// NotFound creates a new not found (404) error.
func NotFound(message string) *Error {
	return NewError(message, http.StatusNotFound)
}

// InternalServerError creates a new internal server error (500).
func InternalServerError(message string) *Error {
	return NewError(message, http.StatusInternalServerError)
}

// BadGateway creates a new bad gateway (502) error.
func BadGateway(message string) *Error {
	return NewError(message, http.StatusBadGateway)
}

// ErrorResponse error response annotated with request context.
type ErrorResponse struct {
	ErrorID    string `json:"errorId,omitempty"`
	Message    string `json:"message,omitempty"`
	StatusCode int    `json:"status,omitempty"`
	Path       string `json:"path,omitempty"`
	RequestID  string `json:"requestId,omitempty"`
}

// NewErrorResponse creates a new error response based on an error an gin context.
func NewErrorResponse(c *gin.Context, err error) ErrorResponse {
	var httpError *Error
	switch err.(type) {
	case *Error:
		httpError = err.(*Error)
		break
	default:
		httpError = InternalServerError(err.Error())
		break
	}

	return ErrorResponse{
		ErrorID:    httpError.ID,
		Message:    httpError.Message,
		StatusCode: httpError.StatusCode,
		Path:       c.Request.URL.Path,
		RequestID:  GetRequestID(c),
	}
}

// getFirstError returns the first error in the gin.Context, nil if not present.
func getFirstError(c *gin.Context) error {
	allErrors := c.Errors
	if len(allErrors) == 0 {
		return nil
	}
	return allErrors[0].Err
}

func logError(c *gin.Context, err ErrorResponse) {
	if err.StatusCode < 500 {
		return
	}

	errLog.Error(err.Message,
		zap.Int("status", err.StatusCode),
		zap.String("errorId", err.ErrorID),
		zap.String("requestId", err.RequestID))
}

func sendError(c *gin.Context, err ErrorResponse) {
	c.AbortWithStatusJSON(err.StatusCode, err)
}
