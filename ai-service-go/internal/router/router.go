package router

import (
	"net/http"
	"time"

	"ai-service-go/config"
	"ai-service-go/internal/handler"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(cfg *config.Config, jobsH *handler.JobsHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"env":       cfg.Env,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if cfg.Env != "production" {
		r.GET("/jobs/:job_id", jobsH.GetJob)
	}
	r.POST("/jobs/analyze", jobsH.Analyze)
	r.POST("/jobs/crawl", jobsH.Crawl)
	r.POST("/jobs/report", jobsH.Report)

	return r
}
