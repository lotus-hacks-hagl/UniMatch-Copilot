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
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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
	userRepo := repository.NewUserRepository(db)
	caseRepo := repository.NewCaseRepository(db)
	uniRepo := repository.NewUniversityRepository(db)
	actRepo := repository.NewActivityRepository(db)
	dashRepo := repository.NewDashboardRepository(db)
	studentRepo := repository.NewStudentRepository(db)

	// 5. Services
	authSvc := service.NewAuthService(userRepo, cfg)
	caseSvc := service.NewCaseService(db, caseRepo, uniRepo, actRepo, aiClient, cfg)
	uniSvc := service.NewUniversityService(db, uniRepo, actRepo, aiClient, cfg)
	dashSvc := service.NewDashboardService(dashRepo, actRepo)
	adminSvc := service.NewAdminService(userRepo)
	studentSvc := service.NewStudentService(studentRepo)

	// 6. Handlers
	authH := handler.NewAuthHandler(authSvc)
	casesH := handler.NewCasesHandler(caseSvc)
	uniH := handler.NewUniversitiesHandler(uniSvc)
	dashH := handler.NewDashboardHandler(dashSvc)
	internalH := handler.NewInternalHandler(caseSvc, uniSvc)
	adminH := handler.NewAdminHandler(adminSvc)
	studentH := handler.NewStudentHandler(studentSvc)

	// 7. Router
	r := router.SetupRouter(cfg, authH, casesH, uniH, dashH, internalH, adminH, studentH)

	log.Printf("🚀 UniMatch-BE running on :%s (env: %s)", cfg.Port, cfg.Env)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
