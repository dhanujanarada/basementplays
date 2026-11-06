package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestETagMatches(t *testing.T) {
	const etag = `"abc123"`

	tests := []struct {
		header string
		want   bool
	}{
		{"", false},
		{etag, true},
		{"W/" + etag, true},
		{`"other", ` + etag, true},
		{`W/"other", W/` + etag, true},
		{"*", true},
		{`"other"`, false},
		{`"abc"`, false},
	}

	for _, tt := range tests {
		if got := etagMatches(tt.header, etag); got != tt.want {
			t.Errorf("etagMatches(%q) = %v, want %v", tt.header, got, tt.want)
		}
	}
}

func TestWriteETagJSON(t *testing.T) {
	api := &publicAPI{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	body := map[string]any{
		"show_id": "show-1",
		"seats":   []string{"S001", "S002"},
	}

	first := httptest.NewRecorder()
	api.writeETagJSON(first, httptest.NewRequest(http.MethodGet, "/", nil), body)

	if first.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", first.Code)
	}
	if got := first.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("content type = %q", got)
	}
	etag := first.Header().Get("ETag")
	if etag == "" {
		t.Fatal("ETag header is empty")
	}
	if got := first.Header().Get("Cache-Control"); got != "no-cache" {
		t.Errorf("cache control = %q, want no-cache", got)
	}

	var decoded map[string]any
	if err := json.Unmarshal(first.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if decoded["show_id"] != "show-1" {
		t.Errorf("decoded body = %v", decoded)
	}

	notModified := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-None-Match", etag)
	api.writeETagJSON(notModified, req, body)

	if notModified.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", notModified.Code)
	}
	if notModified.Body.Len() != 0 {
		t.Errorf("304 body = %q, want empty", notModified.Body.String())
	}
	if got := notModified.Header().Get("ETag"); got != etag {
		t.Errorf("304 ETag = %q, want %q", got, etag)
	}
}

func TestWriteETagJSONChangesWithBody(t *testing.T) {
	api := &publicAPI{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}

	first := httptest.NewRecorder()
	api.writeETagJSON(first, httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{"status": "available"})

	second := httptest.NewRecorder()
	api.writeETagJSON(second, httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{"status": "held"})

	if first.Header().Get("ETag") == second.Header().Get("ETag") {
		t.Fatal("different bodies produced the same ETag")
	}
}
