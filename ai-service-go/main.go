package main

import (
	"log"

	"ai-service-go/config"
	_ "ai-service-go/docs"
	"ai-service-go/internal/handler"
	"ai-service-go/internal/provider"
	"ai-service-go/internal/router"
	"ai-service-go/internal/service"
)

// @title UniMatch AI Service API
// @version 1.0
// @description Simple AI orchestration service for UniMatch using Exa and TinyFish with heuristic fallback.
// @host localhost:9000
// @BasePath /
func main() {
	cfg := config.Load()

	exaClient := provider.NewExaClient(cfg)
	tinyfishClient := provider.NewTinyfishClient(cfg)
	openAIClient := provider.NewOpenAIClient(cfg)
	jobSvc := service.NewJobService(cfg, exaClient, tinyfishClient, openAIClient)
	jobHandler := handler.NewJobsHandler(jobSvc)

	r := router.SetupRouter(cfg, jobHandler)

	log.Printf("ai-service-go running on :%s (%s)", cfg.Port, cfg.Env)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
