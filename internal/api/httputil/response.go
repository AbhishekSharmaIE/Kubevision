package httputil

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID returns X-Request-ID or generates one.
func RequestID(c *gin.Context) string {
	id := c.GetHeader("X-Request-ID")
	if id == "" {
		id = uuid.NewString()
		c.Header("X-Request-ID", id)
	}
	return id
}

// ErrorJSON writes a consistent error body.
func ErrorJSON(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error":     message,
		"requestId": RequestID(c),
	})
}

// DataJSON wraps a successful payload.
func DataJSON(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{
		"data":      data,
		"requestId": RequestID(c),
	})
}

// DataJSONWithMeta adds pagination or extra fields alongside data.
func DataJSONWithMeta(c *gin.Context, status int, data any, meta gin.H) {
	body := gin.H{
		"data":      data,
		"requestId": RequestID(c),
	}
	for k, v := range meta {
		body[k] = v
	}
	c.JSON(status, body)
}

// AbortWithErrorJSON aborts the chain with JSON error.
func AbortWithErrorJSON(c *gin.Context, status int, message string) {
	ErrorJSON(c, status, message)
	c.Abort()
}

// Statuses reused by handlers.
const (
	StatusUnprocessable = http.StatusUnprocessableEntity
)
