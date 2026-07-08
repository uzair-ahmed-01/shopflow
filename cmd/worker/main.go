package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"shopflow/internal/config"
	"shopflow/internal/db"
	"shopflow/internal/repository"
	"shopflow/internal/worker"
)

func main() {
	// Configure zerolog global configurations
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(os.Stdout)

	log.Info().Msg("ShopFlow Worker process starting...")

	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// 2. Establish database connection pool
	dbPool, err := db.NewConnectionPool(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer func() {
		log.Info().Msg("Closing database connection pool...")
		if err := dbPool.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing database connection pool")
		}
	}()

	// 3. Initialize layers
	orderRepo := repository.NewOrderRepository(dbPool)
	orderProcessor := worker.NewOrderProcessor(orderRepo, 3) // 3 background workers

	// Create context that can be cancelled during graceful shutdown
	processorCtx, cancelProcessor := context.WithCancel(context.Background())
	defer cancelProcessor()
	orderProcessor.Start(processorCtx)

	// 4. Graceful shutdown handler
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	sig := <-shutdownChan
	log.Info().Msgf("Received signal %v, shutting down worker...", sig)

	log.Info().Msg("Stopping background order processor...")
	cancelProcessor()
	orderProcessor.Stop()

	log.Info().Msg("Worker process gracefully stopped.")
}
