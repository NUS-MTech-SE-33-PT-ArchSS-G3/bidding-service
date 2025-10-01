package server

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	swgui "github.com/swaggest/swgui/v4"

	"go.uber.org/zap"
)

type Config struct {
	IsEnabled   bool   `json:"IsEnabled"`
	Title       string `json:"Title"`
	OpenApiName string `json:"OpenApiName"`
}

func Serve(getSwagger func() (*openapi3.T, error), r *gin.Engine, cfg *Config, log *zap.Logger) {
	if cfg.IsEnabled {
		log.Info("Serving Swagger UI")
		r.StaticFile("/swagger-spec/"+cfg.OpenApiName+".yaml", "./swagger/"+cfg.OpenApiName+".yaml")

		r.GET("/swagger-spec/"+cfg.OpenApiName+".json", func(c *gin.Context) {
			swagger, err := getSwagger()
			if err != nil {
				c.JSON(500, gin.H{"error": "cannot load control swagger"})
				return
			}
			c.JSON(200, swagger)
		})

		r.GET("/swagger/"+cfg.OpenApiName+"/*any", gin.WrapH(swgui.NewHandler(
			cfg.Title,
			"/swagger-spec/"+cfg.OpenApiName+".json",
			"/swagger/"+cfg.OpenApiName+"/",
		)))
	} else {
		log.Info("Not serving Swagger UI")
	}
}
