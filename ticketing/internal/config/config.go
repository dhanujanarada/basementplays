package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort    = 8080
	minSecretChars = 16
)

// Config is the complete runtime configuration for the ticketing service.
// Required values fail fast; optional values have documented defaults.
type Config struct {
	DatabaseURL string
	Port        int
	LogLevel    string

	InternalToken string
	QRSecret      string
	PublicBaseURL string

	HoldMinutes         int
	CheckoutHoldMinutes int
	ManualHoldHours     int
	MaxSeatsPerOrder    int
	MaxPlanSeats        int

	CORSAllowedOrigins []string

	PayHereMerchantID     string
	PayHereMerchantSecret string
	PayHereSandbox        bool
	PayHereAllowLive      bool

	RunMigrationsOnStart bool
	ShutdownTimeout      time.Duration
}

// Load reads configuration from the process environment.
func Load() (Config, error) {
	return LoadFrom(os.LookupEnv)
}

// LoadFrom is the testable form of Load. get must behave like os.LookupEnv.
func LoadFrom(get func(string) (string, bool)) (Config, error) {
	var problems []string

	cfg := Config{
		DatabaseURL:           value(get, "DATABASE_URL"),
		Port:                  intValue(get, "PORT", defaultPort, &problems),
		LogLevel:              strings.ToLower(value(get, "LOG_LEVEL")),
		InternalToken:         value(get, "INTERNAL_TOKEN"),
		QRSecret:              value(get, "QR_SECRET"),
		PublicBaseURL:         strings.TrimRight(value(get, "PUBLIC_BASE_URL"), "/"),
		HoldMinutes:           intValue(get, "HOLD_MINUTES", 10, &problems),
		CheckoutHoldMinutes:   intValue(get, "CHECKOUT_HOLD_MINUTES", 15, &problems),
		ManualHoldHours:       intValue(get, "MANUAL_HOLD_HOURS", 24, &problems),
		MaxSeatsPerOrder:      intValue(get, "MAX_SEATS_PER_ORDER", 6, &problems),
		MaxPlanSeats:          intValue(get, "MAX_PLAN_SEATS", 500, &problems),
		PayHereMerchantID:     value(get, "PAYHERE_MERCHANT_ID"),
		PayHereMerchantSecret: value(get, "PAYHERE_MERCHANT_SECRET"),
		PayHereSandbox:        boolValue(get, "PAYHERE_SANDBOX", true, &problems),
		PayHereAllowLive:      boolValue(get, "PAYHERE_ALLOW_LIVE", false, &problems),
		RunMigrationsOnStart:  boolValue(get, "RUN_MIGRATIONS_ON_START", true, &problems),
		ShutdownTimeout:       durationValue(get, "SHUTDOWN_TIMEOUT", 10*time.Second, &problems),
	}

	if cfg.DatabaseURL == "" {
		problems = append(problems, "DATABASE_URL is required")
	} else if strings.Contains(cfg.DatabaseURL, "://") {
		u, err := url.Parse(cfg.DatabaseURL)
		if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
			problems = append(problems, "DATABASE_URL must be a postgres:// URL or a libpq keyword string")
		}
	}

	if len(cfg.InternalToken) < minSecretChars {
		problems = append(problems, fmt.Sprintf("INTERNAL_TOKEN must be at least %d characters", minSecretChars))
	}
	if len(cfg.QRSecret) < minSecretChars {
		problems = append(problems, fmt.Sprintf("QR_SECRET must be at least %d characters", minSecretChars))
	}
	if cfg.PublicBaseURL == "" {
		problems = append(problems, "PUBLIC_BASE_URL is required")
	} else if err := validateHTTPURL(cfg.PublicBaseURL); err != nil {
		problems = append(problems, fmt.Sprintf("PUBLIC_BASE_URL %s", err))
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		problems = append(problems, "PORT must be between 1 and 65535")
	}
	switch cfg.LogLevel {
	case "", "debug", "info", "warn", "error":
	default:
		problems = append(problems, "LOG_LEVEL must be one of debug, info, warn, error")
	}

	origins, originProblems := parseOrigins(value(get, "CORS_ALLOWED_ORIGINS"))
	cfg.CORSAllowedOrigins = origins
	problems = append(problems, originProblems...)

	if !cfg.PayHereSandbox && !cfg.PayHereAllowLive {
		problems = append(problems, "PAYHERE_SANDBOX=false requires PAYHERE_ALLOW_LIVE=true")
	}

	if len(problems) > 0 {
		return Config{}, fmt.Errorf("invalid configuration: %s", strings.Join(problems, "; "))
	}

	return cfg, nil
}

// SlogLevel maps LOG_LEVEL to a slog level.
func (c Config) SlogLevel() slog.Level {
	switch c.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func value(get func(string) (string, bool), key string) string {
	if raw, ok := get(key); ok {
		return strings.TrimSpace(raw)
	}
	return ""
}

func intValue(get func(string) (string, bool), key string, def int, problems *[]string) int {
	raw := value(get, key)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		*problems = append(*problems, fmt.Sprintf("%s must be a positive integer", key))
		return def
	}
	return n
}

func boolValue(get func(string) (string, bool), key string, def bool, problems *[]string) bool {
	raw := value(get, key)
	if raw == "" {
		return def
	}
	b, err := strconv.ParseBool(raw)
	if err != nil {
		*problems = append(*problems, fmt.Sprintf("%s must be true or false", key))
		return def
	}
	return b
}

func durationValue(get func(string) (string, bool), key string, def time.Duration, problems *[]string) time.Duration {
	raw := value(get, key)
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		*problems = append(*problems, fmt.Sprintf("%s must be a positive duration (e.g. 10s)", key))
		return def
	}
	return d
}

func validateHTTPURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("is not a valid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("must use http or https")
	}
	if u.Host == "" {
		return fmt.Errorf("must include a host")
	}
	return nil
}

func parseOrigins(raw string) ([]string, []string) {
	var (
		origins  []string
		problems []string
	)

	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if part == "*" {
			problems = append(problems, "CORS_ALLOWED_ORIGINS must list explicit origins, not *")
			continue
		}
		u, err := url.Parse(part)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || (u.Path != "" && u.Path != "/") {
			problems = append(problems, fmt.Sprintf("CORS_ALLOWED_ORIGINS contains invalid origin %q", part))
			continue
		}
		origins = append(origins, u.Scheme+"://"+u.Host)
	}

	return origins, problems
}
