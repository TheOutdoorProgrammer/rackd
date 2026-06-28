// Package api wires the HTTP handlers, middleware, and embedded SPA together.
package api

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/TheOutdoorProgrammer/rackd/internal/config"
	"github.com/TheOutdoorProgrammer/rackd/internal/db"
	"github.com/TheOutdoorProgrammer/rackd/internal/specs"
	"github.com/TheOutdoorProgrammer/rackd/internal/vault"
)

const sessionAuthKey = "authenticated"

// Server holds the HTTP dependencies and the composed handler.
type Server struct {
	cfg      config.Config
	vault    *vault.Vault
	lockout  *vault.Lockout
	sessions *scs.SessionManager
	store    *db.Store
	specs    *specs.Client
	handler  http.Handler
}

// NewServer builds the router and returns a ready http.Handler.
func NewServer(cfg config.Config, v *vault.Vault, lockout *vault.Lockout, sessions *scs.SessionManager, store *db.Store, sp *specs.Client) *Server {
	s := &Server{cfg: cfg, vault: v, lockout: lockout, sessions: sessions, store: store, specs: sp}
	s.handler = s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Server) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(structuredLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(maxBytes(32 << 20)) // generous backstop; uploads enforce a stricter limit

	r.Route("/api", func(r chi.Router) {
		// Public: the frontend needs these before it can authenticate.
		r.Get("/status", s.handleStatus)
		r.Post("/auth/setup", s.handleSetup)
		r.Post("/auth/unlock", s.handleUnlock)

		// Authenticated.
		r.Group(func(r chi.Router) {
			r.Use(s.requireUnlocked)
			r.Get("/me", s.handleMe)
			r.Post("/auth/lock", s.handleLock)
			r.Get("/summary", s.handleSummary)
			r.Get("/search", s.handleSearch)
			r.Get("/report.pdf", s.handleReport)

			r.Route("/firearms", func(r chi.Router) {
				r.Get("/", s.handleListFirearms)
				r.Post("/", s.handleCreateFirearm)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", s.handleGetFirearm)
					r.Put("/", s.handleUpdateFirearm)
					r.Delete("/", s.handleDeleteFirearm)
					r.Get("/ammo", s.handleListFirearmAmmo)
					r.Put("/ammo/{ammoID}", s.handleLinkAmmo)
					r.Delete("/ammo/{ammoID}", s.handleUnlinkAmmo)
				})
			})

			r.Route("/ammo", func(r chi.Router) {
				r.Get("/", s.handleListAmmo)
				r.Post("/", s.handleCreateAmmo)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", s.handleGetAmmo)
					r.Put("/", s.handleUpdateAmmo)
					r.Delete("/", s.handleDeleteAmmo)
					r.Post("/adjust", s.handleAdjustAmmo)
					r.Get("/firearms", s.handleListAmmoFirearms)
				})
			})

			r.Route("/knives", func(r chi.Router) {
				r.Get("/", s.handleListKnives)
				r.Post("/", s.handleCreateKnife)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", s.handleGetKnife)
					r.Put("/", s.handleUpdateKnife)
					r.Delete("/", s.handleDeleteKnife)
				})
			})

			r.Route("/accessories", func(r chi.Router) {
				r.Get("/", s.handleListAccessories) // optional ?firearmId=
				r.Post("/", s.handleCreateAccessory)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", s.handleGetAccessory)
					r.Put("/", s.handleUpdateAccessory)
					r.Delete("/", s.handleDeleteAccessory)
				})
			})

			r.Route("/photos", func(r chi.Router) {
				r.Get("/", s.handleListPhotos)   // ?owner=&id=
				r.Post("/", s.handleUploadPhoto) // ?owner=&id= (multipart)
				r.Get("/{id}", s.handleServePhoto)
				r.Get("/{id}/thumb", s.handleServeThumb)
				r.Put("/{id}/cover", s.handleSetCover)
				r.Post("/{id}/rotate", s.handleRotatePhoto)
				r.Delete("/{id}", s.handleDeletePhoto)
			})

			// Free, key-less firearm spec lookups (Wikipedia + DBpedia), cached.
			r.Route("/specs", func(r chi.Router) {
				r.Get("/search", s.handleSpecsSearch) // ?q=
				r.Get("/page", s.handleSpecsPage)     // ?title=
				r.Get("/cache", s.handleSpecCacheStats)
				r.Delete("/cache", s.handleSpecCacheClear)
			})
		})
	})

	// Everything else falls through to the embedded SPA.
	r.NotFound(s.spaHandler())

	return s.sessions.LoadAndSave(r)
}
