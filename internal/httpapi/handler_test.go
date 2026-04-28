package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"positive/internal/shortener"
)

func TestShortenCreatesShortURL(t *testing.T) {
	var gotURL string
	handler := NewHandler(fakeShortener{
		shortenFunc: func(_ context.Context, rawURL string) (string, error) {
			gotURL = rawURL
			return "abc123", nil
		},
	}, "http://localhost:8080/")

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if gotURL != "https://example.com" {
		t.Fatalf("rawURL = %q, want %q", gotURL, "https://example.com")
	}

	var body struct {
		ShortURL string `json:"short_url"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ShortURL != "http://localhost:8080/abc123" {
		t.Fatalf("short_url = %q, want %q", body.ShortURL, "http://localhost:8080/abc123")
	}
}

func TestShortenInvalidJSON(t *testing.T) {
	handler := NewHandler(fakeShortener{}, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestShortenInvalidURL(t *testing.T) {
	handler := NewHandler(fakeShortener{
		shortenFunc: func(context.Context, string) (string, error) {
			return "", shortener.ErrInvalidURL
		},
	}, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"ftp://example.com"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRedirectFound(t *testing.T) {
	handler := NewHandler(fakeShortener{
		resolveFunc: func(_ context.Context, code string) (string, error) {
			if code != "abc123" {
				t.Fatalf("code = %q, want %q", code, "abc123")
			}
			return "https://example.com", nil
		},
	}, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if location := rec.Header().Get("Location"); location != "https://example.com" {
		t.Fatalf("Location = %q, want %q", location, "https://example.com")
	}
}

func TestRedirectNotFound(t *testing.T) {
	handler := NewHandler(fakeShortener{
		resolveFunc: func(context.Context, string) (string, error) {
			return "", shortener.ErrNotFound
		},
	}, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

type fakeShortener struct {
	shortenFunc func(context.Context, string) (string, error)
	resolveFunc func(context.Context, string) (string, error)
}

func (f fakeShortener) Shorten(ctx context.Context, rawURL string) (string, error) {
	if f.shortenFunc == nil {
		return "", errors.New("unexpected Shorten call")
	}

	return f.shortenFunc(ctx, rawURL)
}

func (f fakeShortener) Resolve(ctx context.Context, code string) (string, error) {
	if f.resolveFunc == nil {
		return "", errors.New("unexpected Resolve call")
	}

	return f.resolveFunc(ctx, code)
}
