package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakePinger struct {
	err error
}

func (f fakePinger) Ping(context.Context) error {
	return f.err
}

func testHandler(pinger Pinger) http.Handler {
	return New(Options{
		Logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		Pinger:  pinger,
		Version: "test",
	})
}

func decodeBody(t *testing.T, body []byte) map[string]string {
	t.Helper()

	var decoded map[string]string
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("response is not JSON: %v (%s)", err, body)
	}
	return decoded
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	testHandler(fakePinger{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", got)
	}
	if body := decodeBody(t, rec.Body.Bytes()); body["status"] != "ok" || body["version"] != "test" {
		t.Errorf("body = %v", body)
	}
}

func TestReadyzReady(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	testHandler(fakePinger{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if body := decodeBody(t, rec.Body.Bytes()); body["status"] != "ready" {
		t.Errorf("body = %v", body)
	}
}

func TestReadyzUnavailable(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	testHandler(fakePinger{err: errors.New("database down")}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if body := decodeBody(t, rec.Body.Bytes()); body["status"] != "unavailable" {
		t.Errorf("body = %v", body)
	}
}

func TestReadyzWithoutPinger(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	New(Options{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func TestRequestIDGenerated(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	testHandler(fakePinger{}).ServeHTTP(rec, req)

	id := rec.Header().Get("X-Request-Id")
	if len(id) != 32 {
		t.Fatalf("generated request id = %q, want 32 hex characters", id)
	}
}

func TestRequestIDEchoedWhenValid(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-Id", "caller-123")

	testHandler(fakePinger{}).ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-Id"); got != "caller-123" {
		t.Fatalf("X-Request-Id = %q, want caller-123", got)
	}
}

func TestRequestIDRejectedWhenInvalid(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-Id", "bad id with spaces")

	testHandler(fakePinger{}).ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-Id"); got == "bad id with spaces" || len(got) != 32 {
		t.Fatalf("X-Request-Id = %q, want a generated id", got)
	}
}

func TestUnknownRouteAndMethod(t *testing.T) {
	handler := testHandler(fakePinger{})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("unknown path status = %d, want 404", rec.Code)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/healthz", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /healthz status = %d, want 405", rec.Code)
	}
}
