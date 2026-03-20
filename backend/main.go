package main

import (
	"log"

	"unimatch-be/config"
	_ "unimatch-be/docs"
	"unimatch-be/internal/handler"
	"unimatch-be/internal/repository"
	"unimatch-be/internal/router"
	"unimatch-be/internal/service"
	"unimatch-be/pkg/client"
	"unimatch-be/pkg/database"
)

// @title UniMatch Copilot API
// @version 1.0
// @description Backend API for UniMatch Copilot AI system
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// 1. Load config
	cfg := config.Load()

	// 2. Connect database + AutoMigrate
	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 3. AI Service client
	aiClient := client.NewAIClient(cfg.AIServiceURL)

	// 4. Repositories
	caseRepo := repository.NewCaseRepository(db)
	uniRepo  := repository.NewUniversityRepository(db)
	actRepo  := repository.NewActivityRepository(db)
	dashRepo := repository.NewDashboardRepository(db)

	// 5. Services
	caseSvc := service.NewCaseService(db, caseRepo, actRepo, aiClient, cfg)
	uniSvc  := service.NewUniversityService(db, uniRepo, actRepo, aiClient, cfg)
	dashSvc := service.NewDashboardService(dashRepo, actRepo)

	// 6. Handlers
	casesH    := handler.NewCasesHandler(caseSvc)
	uniH      := handler.NewUniversitiesHandler(uniSvc)
	dashH     := handler.NewDashboardHandler(dashSvc)
	internalH := handler.NewInternalHandler(caseSvc, uniSvc)

	// 7. Router
	r := router.SetupRouter(casesH, uniH, dashH, internalH)

	log.Printf("🚀 UniMatch-BE running on :%s (env: %s)", cfg.Port, cfg.Env)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
