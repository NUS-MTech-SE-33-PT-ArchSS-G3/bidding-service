package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func MaxBody(byteLimit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// skip methods with no body
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodHead ||
			strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
			c.Next()
			return
		}

		// fast fail if Content-Length exceeds limit
		if cl := c.Request.Header.Get("Content-Length"); cl != "" {
			if n, err := strconv.ParseInt(cl, 10, 64); err == nil && n > byteLimit {
				c.Writer.Header().Set("Connection", "close")
				c.AbortWithStatus(http.StatusRequestEntityTooLarge) // 413
				return
			}
		}

		// enforce max body size
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, byteLimit)

		c.Writer.Header().Set("X-Max-Body-Bytes", strconv.FormatInt(byteLimit, 10))

		c.Next()
	}
}
