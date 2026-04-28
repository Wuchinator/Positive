package shortener

import (
	"context"
	"errors"
	"testing"
)

func TestServiceShortenAndResolve(t *testing.T) {
	ctx := context.Background()
	store := newMemoryStore()
	service := NewService(store)

	code, err := service.Shorten(ctx, " https://example.com/path?q=1 ")
	if err != nil {
		t.Fatalf("Shorten() error = %v", err)
	}
	if code == "" {
		t.Fatal("Shorten() returned empty code")
	}

	originalURL, err := service.Resolve(ctx, code)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if originalURL != "https://example.com/path?q=1" {
		t.Fatalf("Resolve() = %q, want %q", originalURL, "https://example.com/path?q=1")
	}

	repeatedCode, err := service.Shorten(ctx, "https://example.com/path?q=1")
	if err != nil {
		t.Fatalf("Shorten() repeated error = %v", err)
	}
	if repeatedCode != code {
		t.Fatalf("Shorten() repeated code = %q, want %q", repeatedCode, code)
	}
}

func TestServiceShortenRejectsInvalidURLs(t *testing.T) {
	service := NewService(newMemoryStore())

	tests := []struct {
		name   string
		rawURL string
	}{
		{name: "empty", rawURL: ""},
		{name: "without scheme", rawURL: "example.com"},
		{name: "unsupported scheme", rawURL: "ftp://example.com"},
		{name: "without host", rawURL: "https://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Shorten(context.Background(), tt.rawURL)
			if !errors.Is(err, ErrInvalidURL) {
				t.Fatalf("Shorten() error = %v, want %v", err, ErrInvalidURL)
			}
		})
	}
}

func TestServiceResolveUnknownCode(t *testing.T) {
	service := NewService(newMemoryStore())

	_, err := service.Resolve(context.Background(), "unknown")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Resolve() error = %v, want %v", err, ErrNotFound)
	}
}

func TestServiceRetriesWhenGeneratedCodeExists(t *testing.T) {
	ctx := context.Background()
	store := newMemoryStore()
	store.codeExistsFailures = 2
	service := NewService(store)

	code, err := service.Shorten(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Shorten() error = %v", err)
	}
	if code == "" {
		t.Fatal("Shorten() returned empty code")
	}
	if store.saveCalls != 3 {
		t.Fatalf("Save() calls = %d, want 3", store.saveCalls)
	}
}

type memoryStore struct {
	byCode             map[string]string
	byURL              map[string]string
	codeExistsFailures int
	saveCalls          int
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		byCode: make(map[string]string),
		byURL:  make(map[string]string),
	}
}

func (s *memoryStore) Save(_ context.Context, code string, originalURL string) (string, error) {
	s.saveCalls++

	if storedCode, ok := s.byURL[originalURL]; ok {
		return storedCode, nil
	}

	if s.codeExistsFailures > 0 {
		s.codeExistsFailures--
		return "", ErrCodeExists
	}

	if _, ok := s.byCode[code]; ok {
		return "", ErrCodeExists
	}

	s.byCode[code] = originalURL
	s.byURL[originalURL] = code

	return code, nil
}

func (s *memoryStore) Resolve(_ context.Context, code string) (string, error) {
	originalURL, ok := s.byCode[code]
	if !ok {
		return "", ErrNotFound
	}

	return originalURL, nil
}
