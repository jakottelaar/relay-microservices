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
	"github.com/jakottelaar/relay-microservices/shared/errors"
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
		logger.Fatal("Failed to create database pool",
		 zap.Error(err),
		)
	}
	defer pool.Close()

	if err := sonyflake.InitSonyFlake(); err != nil {
		logger.Fatal("Failed to initialize Sonyflake",
		 zap.Error(err),
		)
	}

	r := gin.Default()

	r.Use(errors.ErrorHandler())
	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
	

	repo := internal.NewAuthRepository(pool)
	jwtManager, err := internal.NewJWTManager(cfg, repo)
	if err != nil {
		logger.Fatal("Failed to create JWT manager",
		 zap.Error(err),
		)
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
		logger.Info(
			fmt.Sprintf("Auth service is running on port %s", cfg.Port),
			zap.String("port", cfg.Port),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(
				"Failed to start server",
				zap.Error(err),
			)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal(
			"Failed to gracefully shutdown server",
			zap.Error(err),
		)
	}

	logger.Info("Server exited")
}