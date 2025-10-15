package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const RequestIDKey = "requestId"
const RequestIDHeader = "X-Request-ID"

// RequestID middleware that ensures each request has a unique ID
// If a request does not have a request ID, a new one will be generated
// The request ID will be added to the response header
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

type ctxKey int

const loggerKey ctxKey = iota

// WithRequestLogger adds a logger to the context with request ID pre-populated, use in server setup
func WithRequestLogger(base *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqLog := base.With(zap.String(RequestIDKey, c.GetString(RequestIDKey)))
		c.Set("logger", reqLog)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), loggerKey, reqLog))
		c.Next()
	}
}

// LoggerFrom retrieves a request-scoped logger from the context
// or returns the provided fallback logger if none is found
// Usage: log := middleware.LoggerFrom(c.Request.Context(), h.log)
func LoggerFrom(ctx context.Context, fallback *zap.Logger) *zap.Logger {
	if l, ok := ctx.Value(loggerKey).(*zap.Logger); ok && l != nil {
		return l
	}
	return fallback
}
