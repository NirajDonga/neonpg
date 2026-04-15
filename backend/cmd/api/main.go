package main

import (
	"context"
	"log"
	"net/http"

	"github.com/NirajDonga/dbpods/internal/config"
	database "github.com/NirajDonga/dbpods/internal/db"
	"github.com/NirajDonga/dbpods/internal/handlers"
	"github.com/NirajDonga/dbpods/internal/repository"
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

	// Wire up repositories and handlers.
	userRepo := repository.NewUserRepository(pool)

	authHandler := handlers.NewAuthHandler(cfg, userRepo)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	auth := r.Group("/auth")
	{
		auth.GET("/google", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
	}

	log.Printf("Server starting on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
