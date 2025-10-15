package http

import (
	"kei-services/services/bid-command/openapi"

	"github.com/gin-gonic/gin"
)

func writeProblem(c *gin.Context, status int, typ, title, detail string) {
	p := openapi.ProblemDetails{
		Type:   typ,
		Title:  title,
		Status: status,
	}
	if detail != "" {
		p.Detail = &detail
	}

	c.Header("Content-Type", "application/problem+json")
	c.JSON(status, p)
}

func strDeref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
