package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ContextRequestID is the gin context key for the active request id.
const ContextRequestID = "requestId"

// RequestID ensures X-Request-ID is present on the response (and in context).
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Writer.Header().Set("X-Request-ID", id)
		c.Set(ContextRequestID, id)
		c.Next()
	}
}
