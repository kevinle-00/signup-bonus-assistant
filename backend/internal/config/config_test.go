package config

import "testing"

func TestLoadUsesDefaultAPIAddr(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("API_ADDR", "")
	t.Setenv("PORT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.APIAddr != ":8080" {
		t.Fatalf("APIAddr = %q, want :8080", cfg.APIAddr)
	}
}

func TestLoadUsesRailwayPortWhenAPIAddrUnset(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("API_ADDR", "")
	t.Setenv("PORT", "3000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.APIAddr != ":3000" {
		t.Fatalf("APIAddr = %q, want :3000", cfg.APIAddr)
	}
}

func TestLoadKeepsExplicitAPIAddrOverPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("API_ADDR", ":9000")
	t.Setenv("PORT", "3000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.APIAddr != ":9000" {
		t.Fatalf("APIAddr = %q, want :9000", cfg.APIAddr)
	}
}
