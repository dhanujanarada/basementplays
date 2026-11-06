package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterWindow(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	limiter := newRateLimiter(2, time.Minute)
	limiter.now = func() time.Time { return now }

	for i := 0; i < 2; i++ {
		if ok, _ := limiter.allow("1.2.3.4"); !ok {
			t.Fatalf("request %d denied", i+1)
		}
	}

	ok, retry := limiter.allow("1.2.3.4")
	if ok {
		t.Fatal("third request allowed")
	}
	if retry <= 0 || retry > time.Minute {
		t.Fatalf("retry = %v, want within one minute", retry)
	}

	if ok, _ := limiter.allow("5.6.7.8"); !ok {
		t.Fatal("different key denied")
	}

	now = now.Add(time.Minute)
	if ok, _ := limiter.allow("1.2.3.4"); !ok {
		t.Fatal("request after window denied")
	}
}

func TestWithPublicRateLimit(t *testing.T) {
	limiter := newRateLimiter(1, time.Minute)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := withPublicRateLimit(limiter, next)

	first := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/public/shows/show-1", nil)
	req.RemoteAddr = "203.0.113.9:1234"
	handler.ServeHTTP(first, req)
	if first.Code != http.StatusNoContent {
		t.Fatalf("first status = %d, want 204", first.Code)
	}

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, req)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want 429", second.Code)
	}
	if second.Header().Get("Retry-After") == "" {
		t.Error("Retry-After is empty")
	}

	// Health and internal routes are not part of the public limiter.
	for _, path := range []string{"/healthz", "/v1/internal/seat-plans"} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNoContent {
			t.Errorf("%s status = %d, want 204", path, rec.Code)
		}
	}
}

func TestCORSAllowedOrigin(t *testing.T) {
	handler := withCORS(
		[]string{"https://basementplays.com"},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/public/shows/show-1", nil)
	req.Header.Set("Origin", "https://basementplays.com")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://basementplays.com" {
		t.Errorf("allow origin = %q", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want Origin", got)
	}
}

func TestCORSDisallowedOrigin(t *testing.T) {
	handler := withCORS(
		[]string{"https://basementplays.com"},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/public/shows/show-1", nil)
	req.Header.Set("Origin", "https://evil.example")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("allow origin = %q, want empty", got)
	}
}

func TestCORSPreflight(t *testing.T) {
	handler := withCORS(
		[]string{"https://basementplays.com"},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("preflight reached the next handler")
		}),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/v1/public/shows/show-1/holds", nil)
	req.Header.Set("Origin", "https://basementplays.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("allow methods is empty")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("allow headers is empty")
	}
}

func TestCORSPreflightRejectsUnknownOrigin(t *testing.T) {
	handler := withCORS(
		[]string{"https://basementplays.com"},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("rejected preflight reached the next handler")
		}),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/v1/public/shows/show-1/holds", nil)
	req.Header.Set("Origin", "https://evil.example")
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestCORSOnlyAppliesToPublicPaths(t *testing.T) {
	handler := withCORS(
		[]string{"https://basementplays.com"},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/internal/seat-plans", nil)
	req.Header.Set("Origin", "https://basementplays.com")
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("allow origin = %q, want empty", got)
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "remote addr",
			remoteAddr: "203.0.113.9:1234",
			want:       "203.0.113.9",
		},
		{
			name:       "forwarded for first",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "198.51.100.7, 10.0.0.2"},
			want:       "198.51.100.7",
		},
		{
			name:       "real ip",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"X-Real-IP": "198.51.100.8"},
			want:       "198.51.100.8",
		},
		{
			name:       "invalid forwarded for falls back",
			remoteAddr: "10.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "not-an-ip"},
			want:       "10.0.0.1",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			if got := clientIP(req); got != tt.want {
				t.Fatalf("clientIP = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWithInternalAuth(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := withInternalAuth("a-long-internal-token", next)

	tests := []struct {
		name   string
		header string
		want   int
	}{
		{"missing", "", http.StatusUnauthorized},
		{"wrong scheme", "Basic abc", http.StatusUnauthorized},
		{"wrong token", "Bearer nope", http.StatusUnauthorized},
		{"correct", "Bearer a-long-internal-token", http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/internal/seat-plans", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestWriteRateLimited(t *testing.T) {
	rec := httptest.NewRecorder()
	writeRateLimited(rec, 1500*time.Millisecond)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
	if got := rec.Header().Get("Retry-After"); got != "2" {
		t.Errorf("Retry-After = %q, want 2", got)
	}
}
