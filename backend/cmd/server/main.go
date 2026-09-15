package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/netmon/netmon/internal/api"
	"github.com/netmon/netmon/internal/config"
	"github.com/netmon/netmon/internal/migrations"
	"github.com/netmon/netmon/internal/monitors"
	"github.com/netmon/netmon/internal/scheduler"
	"github.com/netmon/netmon/internal/store"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	logger.Info("starting netmon server",
		zap.String("version", "0.2.0"),
		zap.String("mode", cfg.Server.Mode))

	// Initialize database
	db, err := store.NewStore(cfg.Database.DSN())
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	logger.Info("connected to database")

	// Run database migrations
	logger.Info("running database migrations")
	if err := migrations.RunMigrations(db.DB()); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}
	logger.Info("migrations completed successfully")

	// Initialize monitor factory
	monitorFactory := monitors.NewMonitorFactory()

	// Initialize scheduler
	sched := scheduler.NewScheduler(db, monitorFactory, logger)

	// Initialize API
	apiServer := api.NewAPI(db, logger)

	// Start scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := sched.Start(ctx); err != nil {
		logger.Fatal("Failed to start scheduler", zap.Error(err))
	}

	// Start HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      apiServer.Router(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		logger.Info("starting HTTP server", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("shutting down server")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", zap.Error(err))
	}

	cancel()
	sched.Stop()

	logger.Info("server stopped")
}
