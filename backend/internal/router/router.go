package router

import (
	"net/http"
	"time"

	"unimatch-be/internal/handler"
	"unimatch-be/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(
	casesH    *handler.CasesHandler,
	uniH      *handler.UniversitiesHandler,
	dashH     *handler.DashboardHandler,
	internalH *handler.InternalHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Internal — AI Service callback (not exposed to public, no auth needed in hackathon)
	r.POST("/internal/jobs/done", internalH.JobDone)

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		// Cases routes
		cases := api.Group("/cases")
		{
			cases.POST("", casesH.Create)
			cases.GET("", casesH.List)
			cases.GET("/count", casesH.Count)
			cases.GET("/:id", casesH.GetByID)
			cases.POST("/:id/report", casesH.RequestReport)
		}

		// Universities routes
		unis := api.Group("/universities")
		{
			unis.GET("", uniH.List)
			unis.POST("", uniH.Create)
			unis.POST("/crawl-all", uniH.CrawlAll)
			unis.GET("/crawl-active", uniH.CrawlActiveCount)
		}

		// Dashboard routes
		dash := api.Group("/dashboard")
		{
			dash.GET("/stats", dashH.Stats)
			dash.GET("/cases-by-day", dashH.CasesByDay)
			dash.GET("/escalation-trend", dashH.EscalationTrend)
			dash.GET("/analytics", dashH.Analytics)
		}

		// Activity log
		api.GET("/activity-log", dashH.ActivityLog)
	}

	return r
}
