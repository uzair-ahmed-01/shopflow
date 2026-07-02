package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"shopflow/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const authUserKey contextKey = "authUser"

type AuthUser struct {
	ID    int
	Email string
}

// AuthMiddleware intercepts requests to validate JWT tokens.
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				sendError(w, http.StatusUnauthorized, "missing authorization header", "UNAUTHORIZED")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				sendError(w, http.StatusUnauthorized, "invalid authorization header format", "UNAUTHORIZED")
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
				sendError(w, http.StatusUnauthorized, "invalid or expired token", "UNAUTHORIZED")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				sendError(w, http.StatusUnauthorized, "invalid token claims", "UNAUTHORIZED")
				return
			}

			// Extract claims
			userIDFloat, ok1 := claims["user_id"].(float64)
			email, ok2 := claims["email"].(string)

			if !ok1 || !ok2 {
				sendError(w, http.StatusUnauthorized, "invalid token payload", "UNAUTHORIZED")
				return
			}

			authUser := &AuthUser{
				ID:    int(userIDFloat),
				Email: email,
			}

			// Inject user info into request context
			ctx := context.WithValue(r.Context(), authUserKey, authUser)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext retrieves the authenticated user ID from context.
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	authUser, ok := GetAuthUser(ctx)
	if !ok {
		return 0, false
	}
	return authUser.ID, true
}

// GetAuthUser retrieves the authenticated AuthUser struct from context.
func GetAuthUser(ctx context.Context) (*AuthUser, bool) {
	authUser, ok := ctx.Value(authUserKey).(*AuthUser)
	return authUser, ok
}

// sendError sends an error envelope JSON response directly without depending on internal/handler.
func sendError(w http.ResponseWriter, status int, message string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error": map[string]string{
			"message": message,
			"code":    code,
		},
	})
}
