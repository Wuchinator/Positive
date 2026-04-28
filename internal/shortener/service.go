package shortener

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"
)

const (
	codeLength = 6
)

var (
	ErrInvalidURL     = errors.New("invalid url")
	ErrNotFound       = errors.New("url not found")
	ErrCodeExists     = errors.New("code already exists")
	ErrCodeGeneration = errors.New("failed to generate unique code")
)

type Store interface {
	Save(ctx context.Context, code string, originalURL string) (string, error)
	Resolve(ctx context.Context, code string) (string, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) Shorten(ctx context.Context, rawURL string) (string, error) {
	originalURL, err := validateURL(rawURL)
	if err != nil {
		return "", err
	}

	for range 10 {
		code, err := generateCode()
		if err != nil {
			return "", err
		}

		storedCode, err := s.store.Save(ctx, code, originalURL)
		if errors.Is(err, ErrCodeExists) {
			continue
		}
		if err != nil {
			return "", err
		}

		return storedCode, nil
	}

	return "", ErrCodeGeneration
}

func (s *Service) Resolve(ctx context.Context, code string) (string, error) {
	code = strings.TrimSpace(code)
	if code == "" || strings.Contains(code, "/") {
		return "", ErrNotFound
	}

	return s.store.Resolve(ctx, code)
}

func validateURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", ErrInvalidURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", ErrInvalidURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", ErrInvalidURL
	}

	if parsedURL.Host == "" {
		return "", ErrInvalidURL
	}

	return parsedURL.String(), nil
}

func generateCode() (string, error) {
	bytes := make([]byte, codeLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	code := base64.RawURLEncoding.EncodeToString(bytes)
	if len(code) > codeLength {
		code = code[:codeLength]
	}

	return code, nil
}
