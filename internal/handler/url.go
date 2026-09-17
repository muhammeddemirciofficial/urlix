package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/muhammeddemirciofficial/urlix/internal/service"
)

const maxCreateURLBodySize = 8 << 10

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
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxCreateURLBodySize)
	defer r.Body.Close()

	var req CreateURLRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	result, err := h.urlService.GetURLData(code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		if errors.Is(err, service.ErrURLExpired) {
			http.Error(w, "URL expired", http.StatusGone)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := h.urlService.IncrementClickCount(code); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := h.urlService.RecordClick(result.ID, r.UserAgent(), r.Referer()); err != nil {
		return
	}

	http.Redirect(w, r, result.URL, http.StatusFound)
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

func (h *URLHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	analytics, err := h.urlService.GetAnalytics(code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(analytics); err != nil {
		return
	}
}
