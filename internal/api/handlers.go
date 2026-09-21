package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AbePlays/go-url-shortener-system-design/internal/store"
	"github.com/AbePlays/go-url-shortener-system-design/utils"
)

type Handler struct {
	store   *store.UrlStore
	baseUrl string
}

type ShortenRequest struct {
	Url string `json:"url"`
}

type ShortenResponse struct {
	ShortCode string `json:"shortCode"`
	Url       string `json:"url"`
}

func New(store *store.UrlStore, baseUrl string) *Handler {
	return &Handler{
		store:   store,
		baseUrl: baseUrl,
	}
}

func (handler *Handler) ShortenHandler(w http.ResponseWriter, req *http.Request) {
	var params ShortenRequest
	err := json.NewDecoder(req.Body).Decode(&params)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !utils.IsValidUrl(params.Url) {
		http.Error(w, "Invalid url", http.StatusBadRequest)
		return
	}

	shortCode, err := handler.store.AddUrl(params.Url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	redirectUrl := fmt.Sprintf("%s/%s", handler.baseUrl, shortCode)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ShortenResponse{ShortCode: shortCode, Url: redirectUrl})
}

func (handler *Handler) RedirectHandler(w http.ResponseWriter, req *http.Request) {
	code := req.PathValue("code")

	url, err := handler.store.GetUrl(req.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.Redirect(w, req, url, http.StatusFound)
}
