package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"shopflow/internal/config"
	"shopflow/internal/handler"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	userIDKey    contextKey = "userID"
	userEmailKey contextKey = "userEmail"
)

// AuthMiddleware intercepts requests to validate JWT tokens.
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				handler.SendError(w, http.StatusUnauthorized, "missing authorization header", "UNAUTHORIZED")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				handler.SendError(w, http.StatusUnauthorized, "invalid authorization header format", "UNAUTHORIZED")
				return
			}

			tokenString := parts[1]
			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(cfg.JWTSecret), nil
			})

			if err != nil || !token.Valid {
				handler.SendError(w, http.StatusUnauthorized, "invalid or expired token", "UNAUTHORIZED")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				handler.SendError(w, http.StatusUnauthorized, "invalid token claims", "UNAUTHORIZED")
				return
			}

			// Extract claims
			userIDFloat, ok1 := claims["user_id"].(float64)
			email, ok2 := claims["email"].(string)

			if !ok1 || !ok2 {
				handler.SendError(w, http.StatusUnauthorized, "invalid token payload", "UNAUTHORIZED")
				return
			}

			// Inject user info into request context
			ctx := context.WithValue(r.Context(), userIDKey, int(userIDFloat))
			ctx = context.WithValue(ctx, userEmailKey, email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext retrieves the authenticated user ID from context.
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	val, ok := ctx.Value(userIDKey).(int)
	return val, ok
}

// GetUserEmailFromContext retrieves the authenticated user email from context.
func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(userEmailKey).(string)
	return val, ok
}
