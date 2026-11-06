package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"

	"github.com/dhanujanarada/basementplays/ticketing/internal/service"
)

// Pinger is the readiness dependency (normally *pgxpool.Pool).
type Pinger interface {
	Ping(ctx context.Context) error
}

// Options configures the HTTP handler.
type Options struct {
	Logger  *slog.Logger
	Pinger  Pinger
	Version string

	// InternalToken guards /v1/internal/*. Plans and Shows are required to
	// mount those routes; both must be non-nil.
	InternalToken string
	Plans         *service.Plans
	Shows         *service.Shows
	Orders        *service.Orders

	// Scanner mounts POST /v1/scan/admit and the internal scanner-token
	// endpoints when non-nil.
	Scanner *service.Scanner

	// Availability, Holds, Orders, and Tickets mount the public booking routes
	// when non-nil.
	Availability *service.Availability
	Holds        *service.Holds
	Tickets      *service.Tickets

	// CORSAllowedOrigins is the exact browser origin allowlist for /v1/public/*.
	CORSAllowedOrigins []string
}

type contextKey string

const requestIDKey contextKey = "request_id"

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

// New builds the service handler and mounts whichever route groups are wired.
func New(opts Options) http.Handler {
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}

	mux := http.NewServeMux()

	if opts.Plans != nil && opts.Shows != nil && opts.Orders != nil {
		api := &internalAPI{
			logger:  opts.Logger,
			plans:   opts.Plans,
			shows:   opts.Shows,
			orders:  opts.Orders,
			scanner: opts.Scanner,
		}
		mux.Handle("/v1/internal/", withInternalAuth(
			opts.InternalToken,
			http.StripPrefix("/v1/internal", api.routes()),
		))
	}

	if opts.Scanner != nil {
		api := &scanAPI{logger: opts.Logger, scanner: opts.Scanner}
		mux.Handle("/v1/scan/", http.StripPrefix("/v1/scan", api.routes()))
	}

	if opts.Availability != nil && opts.Holds != nil && opts.Orders != nil && opts.Tickets != nil {
		api := &publicAPI{
			logger:       opts.Logger,
			availability: opts.Availability,
			holds:        opts.Holds,
			orders:       opts.Orders,
			tickets:      opts.Tickets,
			writeLimiter: newRateLimiter(publicWritesPerMinute, rateLimitWindow),
		}
		api.mount(mux)
	}

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"version": opts.Version,
		})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if opts.Pinger == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := opts.Pinger.Ping(ctx); err != nil {
			opts.Logger.Warn("readiness check failed",
				"error", err,
				"request_id", requestIDFrom(r.Context()),
			)
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	var handler http.Handler = mux
	handler = withAccessLog(opts.Logger, handler)
	handler = withRecovery(opts.Logger, handler)
	handler = withPublicRateLimit(
		newRateLimiter(publicRequestsPerMinute, rateLimitWindow),
		handler,
	)
	handler = withCORS(opts.CORSAllowedOrigins, handler)
	handler = withRequestID(handler)

	return handler
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func requestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if !requestIDPattern.MatchString(id) {
			id = newRequestID()
		}

		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}

func newRequestID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "request-id-unavailable"
	}
	return hex.EncodeToString(buf[:])
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func withAccessLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(rec, r)

		if rec.status == 0 {
			rec.status = http.StatusOK
		}

		logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"bytes", rec.bytes,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", requestIDFrom(r.Context()),
			"remote_addr", r.RemoteAddr,
		)
	})
}

func withRecovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("panic recovered",
					"panic", recovered,
					"stack", string(debug.Stack()),
					"request_id", requestIDFrom(r.Context()),
				)
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
