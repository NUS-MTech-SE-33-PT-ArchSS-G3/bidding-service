package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"
const RequestIDHeader = "X-Request-ID"

// todo: propagate request ID to downstream services
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(RequestIDHeader)
		if reqID == "" {
			reqID = uuid.NewString()
		}
		c.Set(RequestIDKey, reqID)

		// add to response header
		c.Writer.Header().Set(RequestIDHeader, reqID)

		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	v, exists := c.Get(RequestIDKey)
	if !exists {
		return ""
	}
	return v.(string)
}
