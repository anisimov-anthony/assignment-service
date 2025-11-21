package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"assignment-service/internal/config"
	httphandler "assignment-service/internal/http"
	"assignment-service/internal/repository/mongodb"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	// Logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	//nolint:errcheck
	defer logger.Sync()

	// Config
	cfg := config.Load()
	logger.Info("application started", zap.Object("config", cfg))

	// Root context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	// Database
	logger.Info("connecting to MongoDB...")
	mongoClient, err := mongodb.NewClient(ctx, cfg.MongoURI, cfg.MongoDB, cfg.MongoConnectTimeout, logger)
	if err != nil {
		logger.Fatal("failed to connect to MongoDB", zap.Error(err))
	}
	logger.Info("successfully connected to MongoDB")

	// Server
	router := httphandler.SetupRouter(mongoClient, logger)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	g.Go(func() error {
		logger.Info("starting HTTP server", zap.String("port", cfg.ServerPort))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", zap.Error(err))
			return err
		}
		logger.Info("HTTP server stopped")
		return nil
	})

	// Graceful Shutdown
	g.Go(func() error {
		<-ctx.Done()

		logger.Info("shutdown signal received, starting graceful shutdown...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.GracefulShutdownTimeout)
		defer cancel()

		// 1. Server
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server forced to shutdown", zap.Error(err))
		} else {
			logger.Info("HTTP server stopped gracefully")
		}

		// 2. Database
		logger.Info("closing MongoDB connection...")
		if err := mongoClient.Close(context.Background()); err != nil {
			logger.Error("error closing MongoDB connection", zap.Error(err))
		} else {
			logger.Info("MongoDB connection closed successfully")
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Fatal("application stopped with error", zap.Error(err))
	}

	logger.Info("application stopped gracefully")
}
