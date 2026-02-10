package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
    Env                string
    Port               string
    DB                 DBConfig
    JWTSecret          string
    AccessTokenExpiry  time.Duration
    RefreshTokenExpiry time.Duration
    MaxSessionsPerUser int
    JWTIssuer          string
}

type DBConfig struct {
	DatabaseUrl string
	MaxConns    int32
	MinConns    int32
	MaxConnLifetime int
	MaxConnIdleTime int
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

	JWTSecret := getEnv("JWT_SECRET", "")
	if JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}
	accessTokenExpiry := time.Duration(getEnvInt("ACCESS_TOKEN_EXPIRY", 900)) * time.Second
	refreshTokenExpiry := time.Duration(getEnvInt("REFRESH_TOKEN_EXPIRY", 168)) * time.Hour
	maxSessionsPerUser := getEnvInt("MAX_SESSIONS_PER_USER", 5)
	jwtIssuer := getEnv("JWT_ISSUER", "relay-auth")
	
	return &Config{
		Port: port,
		DB: DBConfig{
			DatabaseUrl: databaseUrl,
			MaxConns:    int32(maxConns),
			MinConns:    int32(minConns),
			MaxConnLifetime: maxConnLifetime,
			MaxConnIdleTime: maxConnIdleTime,
		},
		JWTSecret: JWTSecret,
		AccessTokenExpiry: accessTokenExpiry,
		RefreshTokenExpiry: refreshTokenExpiry,
		MaxSessionsPerUser: maxSessionsPerUser,
		JWTIssuer: jwtIssuer,
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