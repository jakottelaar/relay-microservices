package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env  string
	Port string
	DB   DBConfig
	Minio MinioConfig
	NatsURL string
}

type DBConfig struct {
	DatabaseUrl     string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime int
	MaxConnIdleTime int
}

type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	DefaultAvatar   string
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
	databaseUrl := getEnv("DATABASE_URL", "")
	if databaseUrl == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}
	maxConns := getEnvInt("DB_MAX_CONNS", 10)
	minConns := getEnvInt("DB_MIN_CONNS", 2)
	maxConnLifetime := getEnvInt("DB_MAX_CONN_LIFETIME", 3600)
	maxConnIdleTime := getEnvInt("DB_MAX_CONN_IDLE_TIME", 1800)

	minioEndpoint := getEnv("MINIO_ENDPOINT", "")
	if minioEndpoint == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT environment variable is required")
	}
	minioAccessKeyID := getEnv("MINIO_ACCESS_KEY_ID", "")
	minioSecretAccessKey := getEnv("MINIO_SECRET_ACCESS_KEY", "")
	minioUseSSL := getEnvBool("MINIO_USE_SSL", false)
	minioDefaultAvatar := getEnv("MINIO_DEFAULT_AVATAR", "")

	natsURL := getEnv("NATS_URL", "nats://localhost:4222")


	return &Config{
		Env:  env,
		Port: port,
		DB: DBConfig{
			DatabaseUrl:     databaseUrl,
			MaxConns:        int32(maxConns),
			MinConns:        int32(minConns),
			MaxConnLifetime: maxConnLifetime,
			MaxConnIdleTime: maxConnIdleTime,
		},
		NatsURL: natsURL,
		Minio: MinioConfig{
			Endpoint:        minioEndpoint,
			AccessKeyID:     minioAccessKeyID,
			SecretAccessKey: minioSecretAccessKey,
			UseSSL:          minioUseSSL,
			DefaultAvatar:   minioDefaultAvatar,
		},
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

func getEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true"
	}
	return fallback
}