package http

import (
	api "kei-services/openapi"

	"github.com/gin-gonic/gin"
)

func writeProblem(c *gin.Context, status int, typ, title, detail string) {
	p := api.ProblemDetails{
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
