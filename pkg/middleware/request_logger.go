package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

func RequestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Get(RequestIDKey)
		c.Next() // process request

		// log request details
		latency := time.Since(start)

		status := c.Writer.Status()
		user, _ := c.Get("username")
		username, _ := user.(string)

		logFields := []zap.Field{
			zap.String("requestID", c.GetHeader(RequestIDHeader)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.String("query", c.Request.URL.RawQuery),
			zap.Int("status", status),
			zap.String("ip", c.ClientIP()),
			zap.String("userAgent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
			zap.String("user", username),
		}

		switch {
		case status >= 500:
			log.Error("HTTP 5XX", logFields...)
		case status >= 400:
			log.Warn("HTTP 4XX", logFields...)
		default:
			log.Info("HTTP request", logFields...)
		}
	}
}
