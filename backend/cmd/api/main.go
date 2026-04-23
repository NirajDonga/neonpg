package main

import (
	"context"
	"log"
	"net/http"

	"github.com/NirajDonga/dbpods/internal/config"
	database "github.com/NirajDonga/dbpods/internal/db"
	"github.com/NirajDonga/dbpods/internal/handlers"
	"github.com/NirajDonga/dbpods/internal/kubernetes"
	"github.com/NirajDonga/dbpods/internal/middleware"
	"github.com/NirajDonga/dbpods/internal/repository"
	"github.com/NirajDonga/dbpods/internal/services"
	"github.com/NirajDonga/dbpods/internal/worker"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	ctx := context.Background()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepository(pool)
	podRepo := repository.NewPodRepository(pool)

	k8sClient, err := kubernetes.NewClient(cfg.KubeConfig, cfg.Namespace)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	podService := services.NewPodService(podRepo, k8sClient)

	// Start the background cleanup worker
	cleanupWorker := worker.NewCleanupWorker(podRepo, k8sClient)
	go cleanupWorker.Start(ctx)

	// Wire up handlers
	authHandler := handlers.NewAuthHandler(cfg, userRepo)
	podHandler := handlers.NewPodHandler(podService)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth := r.Group("/auth")
	{
		auth.GET("/google", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
	}

	// Protected API routes
	api := r.Group("/api")
	api.Use(middleware.AuthRequired(cfg.JWTSecret))
	{
		api.POST("/pods", podHandler.CreatePod)
		api.GET("/pods", podHandler.GetUserPods)
	}

	log.Printf("Server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
