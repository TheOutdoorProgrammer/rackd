package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

func (s *Server) handleSpecsSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeError(w, http.StatusBadRequest, "missing q")
		return
	}
	s.cachedSpec(w, r.Context(), "wiki:search:"+strings.ToLower(q), func(ctx context.Context) (any, error) {
		return s.specs.Search(ctx, q)
	})
}

func (s *Server) handleSpecsPage(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.URL.Query().Get("title"))
	if title == "" {
		writeError(w, http.StatusBadRequest, "missing title")
		return
	}
	s.cachedSpec(w, r.Context(), "wiki:page:"+title, func(ctx context.Context) (any, error) {
		return s.specs.Page(ctx, title)
	})
}

// cachedSpec serves cacheKey from the long-lived cache, or runs produce, caches
// the JSON result, and returns it. Cache hits never touch the upstream sources.
func (s *Server) cachedSpec(w http.ResponseWriter, ctx context.Context, cacheKey string, produce func(context.Context) (any, error)) {
	if data, ok, err := s.store.SpecCacheGet(cacheKey); err == nil && ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Boat-Cache", "hit")
		_, _ = w.Write(data)
		return
	}
	result, err := produce(ctx)
	if err != nil {
		slog.Warn("spec lookup failed", "key", cacheKey, "err", err)
		writeError(w, http.StatusBadGateway, "spec source unavailable")
		return
	}
	data, err := json.Marshal(result)
	if err != nil {
		serverError(w, err)
		return
	}
	_ = s.store.SpecCachePut(cacheKey, data)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Boat-Cache", "miss")
	_, _ = w.Write(data)
}

func (s *Server) handleSpecCacheStats(w http.ResponseWriter, _ *http.Request) {
	n, err := s.store.SpecCacheCount()
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"count": n})
}

func (s *Server) handleSpecCacheClear(w http.ResponseWriter, _ *http.Request) {
	n, err := s.store.SpecCacheClear()
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"cleared": n})
}
