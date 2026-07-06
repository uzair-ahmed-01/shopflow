package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"shopflow/internal/config"
	"shopflow/internal/db"
	"shopflow/internal/handler"
	"shopflow/internal/middleware"
	"shopflow/internal/models"
	"shopflow/internal/repository"
	"shopflow/internal/service"
	"shopflow/internal/worker"
)

func main() {
	// Configure zerolog global configurations
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(os.Stdout)

	log.Info().Msg("ShopFlow API server starting...")

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

	// 3. Initialize layers (Dependency Injection)
	userRepo := repository.NewUserRepository(dbPool)
	authService := service.NewAuthService(userRepo, cfg)
	authHandler := handler.NewAuthHandler(authService)

	categoryRepo := repository.NewCategoryRepository(dbPool)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	productRepo := repository.NewProductRepository(dbPool)
	productService := service.NewProductService(productRepo, categoryRepo)
	productHandler := handler.NewProductHandler(productService)

	cartRepo := repository.NewCartRepository(dbPool)
	cartService := service.NewCartService(cartRepo, productRepo)
	cartHandler := handler.NewCartHandler(cartService)

	orderRepo := repository.NewOrderRepository(dbPool)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	orderProcessor := worker.NewOrderProcessor(orderRepo, 3) // 3 background workers

	// Create context that can be cancelled during graceful shutdown
	processorCtx, cancelProcessor := context.WithCancel(context.Background())
	defer cancelProcessor()
	orderProcessor.Start(processorCtx)

	// Middleware
	authMiddleware := middleware.AuthMiddleware(cfg)
	adminMiddleware := middleware.RequireRole(models.RoleAdmin)

	// 4. Set up router and routes
	router := http.NewServeMux()

	// Registeration and Login routes
	router.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	router.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	router.HandleFunc("POST /api/v1/auth/refresh", authHandler.Refresh)
	router.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)

	// Category routes
	router.HandleFunc("GET /api/v1/categories", categoryHandler.ListCategories)
	router.Handle("POST /api/v1/categories", authMiddleware(adminMiddleware(http.HandlerFunc(categoryHandler.CreateCategory))))

	// Product routes
	router.HandleFunc("GET /api/v1/products", productHandler.ListProducts)
	router.Handle("POST /api/v1/products", authMiddleware(adminMiddleware(http.HandlerFunc(productHandler.CreateProduct))))
	router.Handle("PUT /api/v1/products/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(productHandler.UpdateProduct))))
	router.Handle("DELETE /api/v1/products/{id}", authMiddleware(adminMiddleware(http.HandlerFunc(productHandler.DeleteProduct))))

	// Cart routes
	router.Handle("POST /api/v1/cart/items", authMiddleware(http.HandlerFunc(cartHandler.AddOrUpdateItem)))
	router.Handle("GET /api/v1/cart", authMiddleware(http.HandlerFunc(cartHandler.ViewCart)))
	router.Handle("DELETE /api/v1/cart/items/{productId}", authMiddleware(http.HandlerFunc(cartHandler.RemoveItem)))

	// Order routes
	router.Handle("POST /api/v1/orders", authMiddleware(http.HandlerFunc(orderHandler.PlaceOrder)))
	router.Handle("GET /api/v1/orders", authMiddleware(http.HandlerFunc(orderHandler.ListOrders)))
	router.Handle("GET /api/v1/orders/{id}", authMiddleware(http.HandlerFunc(orderHandler.GetOrder)))

	// Root/healthcheck handler
	router.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"status":"healthy"}`))
	})

	// 5. Configure HTTP server with RequestLoggerMiddleware
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: middleware.RequestLoggerMiddleware(router),
	}

	// 6. Graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		log.Info().Msgf("Server listening on %s...", serverAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErr:
		log.Fatal().Err(err).Msg("Server error")
	case sig := <-shutdownChan:
		log.Info().Msgf("Received signal %v, shutting down server...", sig)

		// Stop background processor first to stop dispatching new jobs
		log.Info().Msg("Stopping background order processor...")
		cancelProcessor()
		orderProcessor.Stop()

		// Create a shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatal().Err(err).Msg("Server shutdown failed")
		}
		log.Info().Msg("Server gracefully stopped.")
	}
}
