package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/muhammeddemirciofficial/urlix/internal/service"
)

type URLHandler struct {
	urlService *service.URLService
}

func NewURLHandler(urlService *service.URLService) *URLHandler {
	return &URLHandler{
		urlService: urlService,
	}
}

type CreateURLRequest struct {
	URL   string `json:"url"`
	Alias string `json:"alias"`
}

type CreateURLResponse struct {
	Code string `json:"code"`
	URL  string `json:"url"`
}

func (h *URLHandler) CreateURL(w http.ResponseWriter, r *http.Request) {
	var req CreateURLRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	var code string

	if req.Alias != "" {
		code, err = h.urlService.CreateURLWithAlias(req.URL, req.Alias)
	} else {
		code, err = h.urlService.CreateURL(req.URL)
	}

	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			http.Error(w, "invalid URL", http.StatusBadRequest)
			return
		}

		if errors.Is(err, service.ErrDuplicateCode) {
			http.Error(w, "alias already exists", http.StatusConflict)
			return
		}

		http.Error(w, "failed to create URL", http.StatusInternalServerError)
		return
	}

	response := CreateURLResponse{
		Code: code,
		URL:  req.URL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	url, err := h.urlService.GetURL(code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := h.urlService.IncrementClickCount(code); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *URLHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	stats, err := h.urlService.GetStats(code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
