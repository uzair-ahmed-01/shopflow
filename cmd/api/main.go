package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shopflow/internal/config"
	"shopflow/internal/db"
	"shopflow/internal/handler"
	"shopflow/internal/repository"
	"shopflow/internal/service"
)

func main() {
	log.Println("ShopFlow API server starting...")

	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Establish database connection pool
	dbPool, err := db.NewConnectionPool(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		log.Println("Closing database connection pool...")
		if err := dbPool.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// 3. Initialize layers (Dependency Injection)
	userRepo := repository.NewUserRepository(dbPool)
	authService := service.NewAuthService(userRepo, cfg)
	authHandler := handler.NewAuthHandler(authService)

	// 4. Set up router and routes
	router := http.NewServeMux()
	router.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	router.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	// Root/healthcheck handler
	router.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"status":"healthy"}`))
	})

	// 5. Configure HTTP server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// 6. Graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Server listening on %s...", serverAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErr:
		log.Fatalf("Server error: %v", err)
	case sig := <-shutdownChan:
		log.Printf("Received signal %v, shutting down server...", sig)

		// Create a shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown failed: %v", err)
		}
		log.Println("Server gracefully stopped.")
	}
}
