package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jakottelaar/relay-microservices/services/users/config"
	"github.com/jakottelaar/relay-microservices/services/users/internal"
	"github.com/jakottelaar/relay-microservices/shared/errors"
	"github.com/jakottelaar/relay-microservices/shared/logger"
	"github.com/jakottelaar/relay-microservices/shared/sonyflake"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log, err := logger.NewLogger(cfg.Env)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("Starting users service",
		zap.String("env", cfg.Env),
		zap.String("port", cfg.Port),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := internal.NewPool(ctx, cfg)
	if err != nil {
		log.Fatal("Failed to create database pool",
			zap.Error(err),
		)
	}
	defer pool.Close()

	
	if err := internal.RunMigrations(cfg.DB.DatabaseUrl); err != nil {
		log.Fatal("Failed to run database migrations",
			zap.Error(err),
		)
	}

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatal("Failed to connect to NATS",
			zap.Error(err),
		)
	}
	defer nc.Close()

	storage, err := internal.NewStorageClient(cfg)
	if err != nil {
		log.Fatal("Failed to initialize MinIO client",
			zap.Error(err),
		)
	}
	
	if err := sonyflake.InitSonyFlake(); err != nil {
		log.Fatal("Failed to initialize Sonyflake",
			zap.Error(err),
		)
	}

	r := gin.Default()
	
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}

	r.Use(errors.ErrorHandler())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	repo := internal.NewUserRepository(pool)
	service := internal.NewUserService(repo, log, storage, cfg)
	handler := internal.NewUserHandler(service)
	
	group := r.Group("/users")
	group.GET("/:id/profile", handler.GetUserProfile)
	
	eventHandler := internal.NewEventHandler(service, nc, log)
    if err := eventHandler.SubscribeToEvents(ctx); err != nil {
        log.Fatal("Failed to subscribe to events", zap.Error(err))
    }


	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	go func() {
		log.Info(
			fmt.Sprintf("Users service is running on port %s", cfg.Port),
			zap.String("port", cfg.Port),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(
				"Failed to start server",
				zap.Error(err),
			)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal(
			"Failed to gracefully shutdown server",
			zap.Error(err),
		)
	}

	log.Info("Server exited")

}