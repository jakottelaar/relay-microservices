package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env  string
	Port string
	NatsURL string
	GuildsGrpcAddr string
}

func LoadConfig() (*Config, error) {
	env := getEnv("ENVIRONMENT", "development")

	if env == "development" {
		err := godotenv.Load(".env.local")
		if err != nil {
			return nil, fmt.Errorf("Error loading .env.local file")
		}
	}

	port := getEnv("PORT", "8080")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	guildsGrpcAddr := getEnv("GUILDS_GRPC_ADDR", "localhost:50051")

	return &Config{
		Env:  env,
		Port: port,
		NatsURL: natsURL,
		GuildsGrpcAddr: guildsGrpcAddr,
	}, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		_, err := fmt.Sscanf(value, "%d", &intValue)
		if err == nil {
			return intValue
		}
	}
	return fallback
}