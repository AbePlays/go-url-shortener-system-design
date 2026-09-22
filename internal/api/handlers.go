package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"text/template"

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

type IndexPageData struct {
	ShortURL string
	Error    string
}

type LinksPageData struct {
	Links []store.Url
}

func New(store *store.UrlStore, baseUrl string) *Handler {
	return &Handler{
		store:   store,
		baseUrl: baseUrl,
	}
}

func (handler *Handler) ShortenFormHandler(w http.ResponseWriter, req *http.Request) {
	var indexTpl = template.Must(template.ParseFiles("web/templates/index.html"))
	err := req.ParseForm()

	if err != nil {
		indexTpl.Execute(w, IndexPageData{Error: "Failed to parse form"})
		return
	}

	url := req.FormValue("url")
	if url == "" {
		indexTpl.Execute(w, IndexPageData{Error: "url is required"})
		return
	}

	if !utils.IsValidUrl(url) {
		indexTpl.Execute(w, IndexPageData{Error: "Invalid url"})
		return
	}

	shortCode, err := handler.store.AddUrl(url)
	if err != nil {
		indexTpl.Execute(w, IndexPageData{Error: err.Error()})
		return
	}

	redirectUrl := fmt.Sprintf("%s/%s", handler.baseUrl, shortCode)
	indexTpl.Execute(w, IndexPageData{ShortURL: redirectUrl})
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

func (handler *Handler) LinksWebHandler(w http.ResponseWriter, req *http.Request) {
	links, err := handler.store.ListUrls(req.Context())
	if err != nil {
		slog.Error("failed to list urls", "error", err)
	}

	linksTpl := template.Must(template.ParseFiles("web/templates/links.html"))
	linksTpl.Execute(w, LinksPageData{Links: links})
}

func (handler *Handler) IndexWebHandler(w http.ResponseWriter, req *http.Request) {
	homeTpl := template.Must(template.ParseFiles("web/templates/index.html"))
	homeTpl.Execute(w, nil)
}

func (handler *Handler) AboutWebHandler(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "web/templates/about.html")
}
