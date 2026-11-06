package httpapi

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	publicRequestsPerMinute = 1200
	publicWritesPerMinute   = 300
	rateLimitWindow         = time.Minute
	rateBucketCleanupAt     = 4096
)

type rateBucket struct {
	start time.Time
	count int
}

// rateLimiter is a small in-memory fixed-window limiter. It is deliberately
// per-instance: the documented deployment runs one Go service.
type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]rateBucket
	limit   int
	window  time.Duration
	now     func() time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		buckets: make(map[string]rateBucket),
		limit:   limit,
		window:  window,
		now:     time.Now,
	}
}

// allow reports whether key may make another request. When it cannot, the
// duration is the time until the current window ends.
func (l *rateLimiter) allow(key string) (bool, time.Duration) {
	if l == nil || l.limit <= 0 {
		return true, 0
	}
	if key == "" {
		key = "unknown"
	}

	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.buckets) > rateBucketCleanupAt {
		for existing, bucket := range l.buckets {
			if now.Sub(bucket.start) >= l.window {
				delete(l.buckets, existing)
			}
		}
	}

	bucket, found := l.buckets[key]
	if !found || now.Sub(bucket.start) >= l.window {
		l.buckets[key] = rateBucket{start: now, count: 1}
		return true, 0
	}

	if bucket.count >= l.limit {
		retry := l.window - now.Sub(bucket.start)
		if retry < time.Second {
			retry = time.Second
		}
		return false, retry
	}

	bucket.count++
	l.buckets[key] = bucket

	return true, 0
}

func withPublicRateLimit(limiter *rateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if limiter == nil || !strings.HasPrefix(r.URL.Path, "/v1/public/") {
			next.ServeHTTP(w, r)
			return
		}

		ok, retry := limiter.allow(clientIP(r))
		if !ok {
			writeRateLimited(w, retry)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeRateLimited(w http.ResponseWriter, retry time.Duration) {
	seconds := int(math.Ceil(retry.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(seconds))
	writeAPIError(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
}

// clientIP prefers the reverse proxy's client headers. The Go service must not
// be exposed directly; the deployment puts it behind the site's proxy.
func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		first := strings.TrimSpace(strings.Split(forwarded, ",")[0])
		if net.ParseIP(first) != nil {
			return first
		}
	}
	if real := strings.TrimSpace(r.Header.Get("X-Real-IP")); net.ParseIP(real) != nil {
		return real
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}

	return strings.TrimSpace(r.RemoteAddr)
}

func withCORS(origins []string, next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		allowed[origin] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/public/") {
			next.ServeHTTP(w, r)
			return
		}

		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		if _, ok := allowed[origin]; !ok {
			if r.Method == http.MethodOptions {
				writeAPIError(w, http.StatusForbidden, "origin_not_allowed", "origin is not allowed")
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Add("Vary", "Origin")
		w.Header().Set("Access-Control-Expose-Headers", "ETag, X-Request-Id, Idempotent-Replayed")

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type, Idempotency-Key, If-None-Match, X-Order-Token, X-Request-Id",
			)
			w.Header().Set("Access-Control-Max-Age", "600")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
