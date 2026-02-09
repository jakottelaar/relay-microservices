package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jakottelaar/relay-microservices/services/auth/config"
	"github.com/jakottelaar/relay-microservices/services/auth/internal"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := internal.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := internal.InitSnowflake(); err != nil {
		log.Fatalf("Failed to initialize Snowflake: %v", err)
	}

	r := gin.Default()

	r.Use(internal.ErrorHandler())
	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	
	jwtManager, err := internal.NewJWTManager(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize JWT manager: %v", err)
	}
	repo := internal.NewAuthRepository(pool)
	service := internal.NewAuthService(repo, jwtManager, cfg)
	handler := internal.NewAuthHandler(service)

	r.POST("/sign-up", handler.SignUp)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}