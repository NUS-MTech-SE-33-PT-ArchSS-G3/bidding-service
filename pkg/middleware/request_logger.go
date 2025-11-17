package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next() // process request

		rid := c.GetString(RequestIDKey)
		if rid == "" {
			rid = c.Writer.Header().Get(RequestIDHeader)
		}

		reqLog := LoggerFrom(c.Request.Context(), log).With(
			zap.String("requestId", rid),
		)

		// log request details
		status := c.Writer.Status()
		user, _ := c.Get("username")
		username, _ := user.(string)

		reqLog = reqLog.With(
			zap.String("requestID", c.GetString(RequestIDKey)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.String("query", c.Request.URL.RawQuery),
			zap.Int("status", status),
			zap.String("ip", c.ClientIP()),
			zap.String("userAgent", c.Request.UserAgent()),
			zap.Duration("latency", time.Since(start)),
			zap.String("user", username))

		switch {
		case status >= 500:
			reqLog.Error("HTTP 5XX")
		case status >= 400:
			reqLog.Warn("HTTP 4XX")
		default:
			reqLog.Info("HTTP request")

		}
	}
}
