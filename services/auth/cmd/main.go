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
	"github.com/jakottelaar/relay-microservices/shared/logger"
	"github.com/jakottelaar/relay-microservices/shared/sonyflake"
	"go.uber.org/zap"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := logger.Init(cfg.Env); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Log.Sync()

	logger.Info("Starting auth service",
		zap.String("env", cfg.Env),
		zap.String("port", cfg.Port),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := internal.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := sonyflake.InitSonyFlake(); err != nil {
		log.Fatalf("Failed to initialize SonyFlake: %v", err)
	}

	r := gin.Default()

	r.Use(internal.ErrorHandler())
	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	

	repo := internal.NewAuthRepository(pool)
	jwtManager, err := internal.NewJWTManager(cfg, repo)
	if err != nil {
		log.Fatalf("Failed to initialize JWT manager: %v", err)
	}
	service := internal.NewAuthService(repo, jwtManager, cfg)
	handler := internal.NewAuthHandler(service)

	r.POST("/sign-up", handler.SignUp)
	r.POST("/sign-in", handler.SignIn)
	r.POST("/refresh", handler.Refresh)
	r.POST("/sign-out", handler.SignOut)

	r.GET("/validate", internal.ValidateMiddleware(jwtManager), handler.Validate)

	protected := r.Group("")
    protected.Use(internal.RequireAuth(jwtManager))
    {
        protected.DELETE("/sessions", handler.RevokeAllSessions)
		protected.DELETE("/sessions/:id", handler.RevokeSessionById)
		protected.GET("/sessions/:id", handler.GetSessionById)
    }

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