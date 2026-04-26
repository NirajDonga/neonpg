package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ProxyPort string
	Namespace string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		ProxyPort: mustGetEnv("PROXY_PORT"),
		Namespace: mustGetEnv("K8S_NAMESPACE"),
	}
}

func mustGetEnv(key string) string {
	value, exists := os.LookupEnv(key)

	if !exists || value == "" {
		log.Fatalf("CRITICAL STARTUP ERROR: Environment variable '%s' is required but not set.", key)
	}

	return value
}
