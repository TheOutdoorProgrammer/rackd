// Command boating-accident is a self-hosted, encrypted inventory for firearms, ammo,
// knives, and accessories. Data is encrypted at rest and unlocked with a PIN.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/alexedwards/scs/v2"

	"github.com/TheOutdoorProgrammer/boating-accident/internal/api"
	"github.com/TheOutdoorProgrammer/boating-accident/internal/config"
	"github.com/TheOutdoorProgrammer/boating-accident/internal/db"
	"github.com/TheOutdoorProgrammer/boating-accident/internal/specs"
	"github.com/TheOutdoorProgrammer/boating-accident/internal/vault"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		return err
	}

	database, err := db.Open(filepath.Join(cfg.DataDir, "boating-accident.db"))
	if err != nil {
		return err
	}
	defer func() { _ = database.Close() }()

	v := vault.New(database, vault.Argon2Params{
		MemoryKiB: cfg.Argon2MemoryKiB,
		Time:      cfg.Argon2Time,
		Threads:   cfg.Argon2Threads,
	})
	lockout := vault.NewLockout()
	store := db.NewStore(database, v)
	specsClient := specs.New()

	sessions := scs.New() // in-memory store: sessions die on restart, re-locking the vault
	sessions.Lifetime = 12 * time.Hour
	sessions.IdleTimeout = time.Hour
	sessions.Cookie.Name = "boating-accident_session"
	sessions.Cookie.HttpOnly = true
	sessions.Cookie.SameSite = http.SameSiteLaxMode
	sessions.Cookie.Secure = !cfg.Dev
	sessions.Cookie.Persist = false

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           api.NewServer(cfg, v, lockout, sessions, store, specsClient),
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("listening", "addr", cfg.Addr, "data_dir", cfg.DataDir, "dev", cfg.Dev)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return httpServer.Shutdown(shutdownCtx)
}
