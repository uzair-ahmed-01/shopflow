package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// responseWriter is a custom wrapper around http.ResponseWriter to capture the HTTP status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newResponseWriter creates a new responseWriter wrapper with a default status code of 200.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code and delegates to the original ResponseWriter.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLoggerMiddleware logs detailed HTTP request statistics in JSON format using zerolog.
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		// Process request
		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		// Create request log event
		event := log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.statusCode).
			Float64("duration_ms", float64(duration.Microseconds())/1000.0).
			Str("ip", r.RemoteAddr)

		// Add user ID if present in request context
		if authUser, ok := GetAuthUser(r.Context()); ok {
			event.Int("user_id", authUser.ID)
			event.Str("role", authUser.Role)
		}

		event.Msg("HTTP Request")
	})
}
