package handler

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

// RouterConfig gathers handler dependencies and authorization middlewares.
type RouterConfig struct {
	AuthHandler     *AuthHandler
	CategoryHandler *CategoryHandler
	ProductHandler  *ProductHandler
	CartHandler     *CartHandler
	OrderHandler    *OrderHandler
	AuthMiddleware  func(http.Handler) http.Handler
	AdminMiddleware func(http.Handler) http.Handler
}

// RegisterRoutes registers all REST endpoints and middleware chain rules to multiplexer router.
func RegisterRoutes(mux *http.ServeMux, cfg RouterConfig) {
	// Registration and Login routes
	mux.HandleFunc("POST /api/v1/auth/register", cfg.AuthHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", cfg.AuthHandler.Login)
	mux.Handle("POST /api/v1/auth/refresh", cfg.AuthMiddleware(http.HandlerFunc(cfg.AuthHandler.Refresh)))
	mux.Handle("POST /api/v1/auth/logout", cfg.AuthMiddleware(http.HandlerFunc(cfg.AuthHandler.Logout)))

	// Category routes
	mux.HandleFunc("GET /api/v1/categories", cfg.CategoryHandler.ListCategories)
	mux.Handle("POST /api/v1/categories", cfg.AuthMiddleware(cfg.AdminMiddleware(http.HandlerFunc(cfg.CategoryHandler.CreateCategory))))

	// Product routes
	mux.HandleFunc("GET /api/v1/products", cfg.ProductHandler.ListProducts)
	mux.Handle("POST /api/v1/products", cfg.AuthMiddleware(cfg.AdminMiddleware(http.HandlerFunc(cfg.ProductHandler.CreateProduct))))
	mux.Handle("PUT /api/v1/products/{id}", cfg.AuthMiddleware(cfg.AdminMiddleware(http.HandlerFunc(cfg.ProductHandler.UpdateProduct))))
	mux.Handle("DELETE /api/v1/products/{id}", cfg.AuthMiddleware(cfg.AdminMiddleware(http.HandlerFunc(cfg.ProductHandler.DeleteProduct))))

	// Cart routes
	mux.Handle("POST /api/v1/cart/items", cfg.AuthMiddleware(http.HandlerFunc(cfg.CartHandler.AddOrUpdateItem)))
	mux.Handle("GET /api/v1/cart", cfg.AuthMiddleware(http.HandlerFunc(cfg.CartHandler.ViewCart)))
	mux.Handle("DELETE /api/v1/cart/items/{productId}", cfg.AuthMiddleware(http.HandlerFunc(cfg.CartHandler.RemoveItem)))

	// Order routes
	mux.Handle("POST /api/v1/orders", cfg.AuthMiddleware(http.HandlerFunc(cfg.OrderHandler.PlaceOrder)))
	mux.Handle("GET /api/v1/orders", cfg.AuthMiddleware(http.HandlerFunc(cfg.OrderHandler.ListOrders)))
	mux.Handle("GET /api/v1/orders/{id}", cfg.AuthMiddleware(http.HandlerFunc(cfg.OrderHandler.GetOrder)))

	// Swagger route
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Root/healthcheck handler
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"status":"healthy"}`))
	})
}
