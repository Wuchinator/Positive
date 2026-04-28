package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"positive/internal/shortener"
)

type Shortener interface {
	Shorten(ctx context.Context, rawURL string) (string, error)
	Resolve(ctx context.Context, code string) (string, error)
}

type Handler struct {
	service Shortener
	baseURL string
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

func NewHandler(service Shortener, baseURL string) http.Handler {
	h := &Handler{
		service: service,
		baseURL: strings.TrimRight(baseURL, "/"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", h.shorten)
	mux.HandleFunc("GET /{code}", h.redirect)

	return mux
}

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	code, err := h.service.Shorten(r.Context(), req.URL)
	if errors.Is(err, shortener.ErrInvalidURL) {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "failed to create short url", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(shortenResponse{
		ShortURL: fmt.Sprintf("%s/%s", h.baseURL, code),
	})
}

func (h *Handler) redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	originalURL, err := h.service.Resolve(r.Context(), code)
	if errors.Is(err, shortener.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "failed to resolve url", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
