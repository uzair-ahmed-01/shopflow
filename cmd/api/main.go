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

	_ "shopflow/docs"
)

// @title ShopFlow API
// @version 1.0
// @description ShopFlow modern backend engineering showcase API in Go.
// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer <your-jwt-token>" to authenticate.
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

	// Middleware
	authMiddleware := middleware.AuthMiddleware(cfg)
	adminMiddleware := middleware.RequireRole(models.RoleAdmin)

	// 4. Set up router and routes
	router := http.NewServeMux()
	handler.RegisterRoutes(router, handler.RouterConfig{
		AuthHandler:     authHandler,
		CategoryHandler: categoryHandler,
		ProductHandler:  productHandler,
		CartHandler:     cartHandler,
		OrderHandler:    orderHandler,
		AuthMiddleware:  authMiddleware,
		AdminMiddleware: adminMiddleware,
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

		// Create a shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatal().Err(err).Msg("Server shutdown failed")
		}
		log.Info().Msg("Server gracefully stopped.")
	}
}
