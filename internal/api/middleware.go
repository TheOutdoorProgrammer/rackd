package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// requireUnlocked rejects requests that lack an authenticated session tied to an
// unlocked vault.
func (s *Server) requireUnlocked(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.sessions.GetBool(r.Context(), sessionAuthKey) || !s.vault.Unlocked() {
			writeError(w, http.StatusUnauthorized, "locked")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// structuredLogger logs one slog line per request.
func structuredLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(ww, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", middleware.GetReqID(r.Context()),
		)
	})
}

// maxBytes limits request body size.
func maxBytes(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}
