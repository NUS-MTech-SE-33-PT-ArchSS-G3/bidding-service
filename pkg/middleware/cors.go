package middleware

import (
	"kei-services/pkg/config"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Cors(cfg *config.Cors, log *zap.Logger) gin.HandlerFunc {
	if !cfg.IsEnabled {
		return func(c *gin.Context) {
			// still handle preflight
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}

			c.Next()
		}
	}

	allowedMethods := strings.Join(cfg.AllowMethods, ",")
	expHeaders := strings.Join(cfg.ExposeHeaders, ",")

	in := func(list []string, v string) bool {
		for _, x := range list {
			if strings.EqualFold(x, v) {
				return true
			}
		}
		return false
	}

	addVary := func(c *gin.Context, v string) { c.Writer.Header().Add("Vary", v) }

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		method := c.Request.Method

		// non-browser / same-origin skip CORS
		if origin == "" {
			c.Next()
			return
		}

		// ws upgrade skip CORS
		if strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
			c.Next()
			return
		}

		originAllowed := in(cfg.AllowOrigins, origin)

		if cfg.AllowCredentials {
			if !originAllowed {
				if method == http.MethodOptions {
					c.AbortWithStatus(http.StatusForbidden)
				} else {
					c.AbortWithStatus(http.StatusForbidden)
				}
				return
			}
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		} else {
			if !originAllowed {
				if method == http.MethodOptions {
					c.AbortWithStatus(http.StatusForbidden)
				} else {
					c.AbortWithStatus(http.StatusForbidden)
				}
				return
			}
			c.Header("Access-Control-Allow-Origin", origin)
		}

		addVary(c, "Origin")
		addVary(c, "Access-Control-Request-Method")
		addVary(c, "Access-Control-Request-Headers")

		if allowedMethods != "" {
			c.Header("Access-Control-Allow-Methods", allowedMethods)
		}

		headers := c.Request.Header.Get("Access-Control-Request-Headers")
		if len(cfg.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ","))
		} else if headers != "" {
			// if not configured, reflect requested headers
			c.Header("Access-Control-Allow-Headers", headers)
		}

		if expHeaders != "" {
			c.Header("Access-Control-Expose-Headers", expHeaders)
		}

		if cfg.AllowMaxAge > 0 {
			age := cfg.AllowMaxAge
			if age > 600 {
				age = 600
			}
			c.Header("Access-Control-Max-Age", strconv.Itoa(age))
		}

		// preflight
		if method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
