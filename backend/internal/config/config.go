package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port               string
	Namespace          string
	DatabaseURL        string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleCallbackURL  string
	JWTSecret          string
	KubeConfig         string
}

func LoadConfig() *AppConfig {
	_ = godotenv.Load()

	return &AppConfig{
		Port:               mustGetEnv("PORT"),
		Namespace:          mustGetEnv("K8S_NAMESPACE"),
		DatabaseURL:        mustGetEnv("DATABASE_URL"),
		GoogleClientID:     mustGetEnv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: mustGetEnv("GOOGLE_CLIENT_SECRET"),
		GoogleCallbackURL:  mustGetEnv("GOOGLE_CALLBACK_URL"),
		JWTSecret:          mustGetEnv("JWT_SECRET"),
		KubeConfig:         mustGetEnv("KUBECONFIG"),
	}
}

func mustGetEnv(key string) string {
	value, exists := os.LookupEnv(key)

	if !exists || value == "" {
		log.Fatalf("CRITICAL STARTUP ERROR: Environment variable '%s' is required but not set.", key)
	}

	return value
}
