// Package config loads runtime configuration from RACKD_* environment variables.
// Every setting has a sensible default so rackd runs with zero configuration.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config is the fully-resolved runtime configuration.
type Config struct {
	Addr    string // RACKD_ADDR — listen address (default ":8080")
	DataDir string // RACKD_DATA_DIR — directory for the SQLite DB + encrypted uploads (default "./data")
	Dev     bool   // RACKD_DEV — relaxes the Secure cookie flag for http://localhost development

	// Argon2id cost parameters, used only when a brand-new vault is created.
	// They are persisted per-vault, so changing them later never locks an
	// existing vault out.
	Argon2MemoryKiB uint32 // derived from RACKD_ARGON2_MEMORY_MB
	Argon2Time      uint32 // RACKD_ARGON2_TIME
	Argon2Threads   uint8  // RACKD_ARGON2_THREADS
}

// Load reads configuration from the environment.
func Load() (Config, error) {
	c := Config{
		Addr:            env("RACKD_ADDR", ":8080"),
		DataDir:         env("RACKD_DATA_DIR", "./data"),
		Dev:             envBool("RACKD_DEV", false),
		Argon2MemoryKiB: uint32(envInt("RACKD_ARGON2_MEMORY_MB", 256)) * 1024,
		Argon2Time:      uint32(envInt("RACKD_ARGON2_TIME", 4)),
		Argon2Threads:   uint8(envInt("RACKD_ARGON2_THREADS", 4)),
	}
	if c.Argon2MemoryKiB == 0 || c.Argon2Time == 0 || c.Argon2Threads == 0 {
		return Config{}, fmt.Errorf("config: argon2 parameters must be non-zero")
	}
	return c, nil
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
