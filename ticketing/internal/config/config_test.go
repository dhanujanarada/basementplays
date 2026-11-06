package config

import (
	"strings"
	"testing"
	"time"
)

func getter(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		v, ok := values[key]
		return v, ok
	}
}

func validEnv() map[string]string {
	return map[string]string{
		"DATABASE_URL":    "postgres:///ticketing_dev?host=/var/run/postgresql",
		"INTERNAL_TOKEN":  "test-internal-token",
		"QR_SECRET":       "test-qr-secret-value",
		"PUBLIC_BASE_URL": "http://localhost:3000",
	}
}

func TestLoadFromDefaults(t *testing.T) {
	cfg, err := LoadFrom(getter(validEnv()))
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	checks := []struct {
		name string
		got  int
		want int
	}{
		{"Port", cfg.Port, 8080},
		{"HoldMinutes", cfg.HoldMinutes, 10},
		{"CheckoutHoldMinutes", cfg.CheckoutHoldMinutes, 15},
		{"ManualHoldHours", cfg.ManualHoldHours, 24},
		{"MaxSeatsPerOrder", cfg.MaxSeatsPerOrder, 6},
		{"MaxPlanSeats", cfg.MaxPlanSeats, 500},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}

	if !cfg.PayHereSandbox {
		t.Error("PayHereSandbox = false, want true")
	}
	if cfg.PayHereAllowLive {
		t.Error("PayHereAllowLive = true, want false")
	}
	if !cfg.RunMigrationsOnStart {
		t.Error("RunMigrationsOnStart = false, want true")
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %s, want 10s", cfg.ShutdownTimeout)
	}
	if cfg.LogLevel != "" || cfg.SlogLevel().String() != "INFO" {
		t.Errorf("default log level = %q (%s), want INFO", cfg.LogLevel, cfg.SlogLevel())
	}
}

func TestLoadFromMissingRequired(t *testing.T) {
	for _, key := range []string{"DATABASE_URL", "INTERNAL_TOKEN", "QR_SECRET", "PUBLIC_BASE_URL"} {
		t.Run(key, func(t *testing.T) {
			env := validEnv()
			delete(env, key)

			_, err := LoadFrom(getter(env))
			if err == nil {
				t.Fatalf("LoadFrom() with %s unset returned nil error", key)
			}
			if !strings.Contains(err.Error(), key) {
				t.Fatalf("error %q does not mention %s", err, key)
			}
		})
	}
}

func TestLoadFromRejectsShortSecrets(t *testing.T) {
	env := validEnv()
	env["INTERNAL_TOKEN"] = "short"
	env["QR_SECRET"] = "short"

	_, err := LoadFrom(getter(env))
	if err == nil {
		t.Fatal("LoadFrom() accepted short secrets")
	}
	if !strings.Contains(err.Error(), "INTERNAL_TOKEN") || !strings.Contains(err.Error(), "QR_SECRET") {
		t.Fatalf("error %q does not mention both secrets", err)
	}
}

func TestLoadFromInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		key  string
		val  string
		want string
	}{
		{"bad port", "PORT", "0", "PORT"},
		{"bad int", "HOLD_MINUTES", "soon", "HOLD_MINUTES"},
		{"negative int", "MAX_SEATS_PER_ORDER", "-1", "MAX_SEATS_PER_ORDER"},
		{"bad bool", "PAYHERE_SANDBOX", "maybe", "PAYHERE_SANDBOX"},
		{"bad duration", "SHUTDOWN_TIMEOUT", "later", "SHUTDOWN_TIMEOUT"},
		{"bad base url", "PUBLIC_BASE_URL", "localhost:3000", "PUBLIC_BASE_URL"},
		{"bad log level", "LOG_LEVEL", "verbose", "LOG_LEVEL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := validEnv()
			env[tt.key] = tt.val

			_, err := LoadFrom(getter(env))
			if err == nil {
				t.Fatalf("LoadFrom() accepted %s=%q", tt.key, tt.val)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error %q does not mention %s", err, tt.want)
			}
		})
	}
}

func TestLoadFromPayHereLiveGuard(t *testing.T) {
	env := validEnv()
	env["PAYHERE_SANDBOX"] = "false"

	if _, err := LoadFrom(getter(env)); err == nil {
		t.Fatal("LoadFrom() allowed live PayHere without PAYHERE_ALLOW_LIVE=true")
	}

	env["PAYHERE_ALLOW_LIVE"] = "true"
	cfg, err := LoadFrom(getter(env))
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if !cfg.PayHereAllowLive || cfg.PayHereSandbox {
		t.Fatalf("PayHere flags = sandbox:%v allowLive:%v", cfg.PayHereSandbox, cfg.PayHereAllowLive)
	}
}

func TestLoadFromCORSOrigins(t *testing.T) {
	env := validEnv()
	env["CORS_ALLOWED_ORIGINS"] = "https://example.com, http://localhost:3000/, https://example.com"

	cfg, err := LoadFrom(getter(env))
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}

	want := []string{"https://example.com", "http://localhost:3000", "https://example.com"}
	if len(cfg.CORSAllowedOrigins) != len(want) {
		t.Fatalf("origins = %v, want %v", cfg.CORSAllowedOrigins, want)
	}
	for i := range want {
		if cfg.CORSAllowedOrigins[i] != want[i] {
			t.Errorf("origins[%d] = %q, want %q", i, cfg.CORSAllowedOrigins[i], want[i])
		}
	}

	env["CORS_ALLOWED_ORIGINS"] = "*"
	if _, err := LoadFrom(getter(env)); err == nil {
		t.Fatal("LoadFrom() accepted a wildcard origin")
	}
}
