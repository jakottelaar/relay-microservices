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
	"github.com/jakottelaar/relay-microservices/services/gateway/config"
	"github.com/jakottelaar/relay-microservices/services/gateway/internal"
	"github.com/jakottelaar/relay-microservices/shared/errors"
	"github.com/jakottelaar/relay-microservices/shared/logger"
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

	log.Info("Starting guilds service",
		zap.String("env", cfg.Env),
		zap.String("port", cfg.Port),
	)

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatal("Failed to connect to NATS",
			zap.Error(err),
		)
	}
	defer nc.Close()

	r := gin.Default()

	r.Use(errors.ErrorHandler())
	r.Use(internal.UserContext())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	hub := internal.NewHub(log)
	go hub.Run()

	handler := internal.NewHandler(hub, log)

	r.GET("/ws", handler.ServeWS)

	
	eventHandler := internal.NewEventHandler(nc, hub, log)
	if err := eventHandler.Subscribe(); err != nil {
		log.Fatal("failed to subscribe to NATS", zap.Error(err))
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	go func() {
		log.Info(
			fmt.Sprintf("Guilds service is running on port %s", cfg.Port),
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