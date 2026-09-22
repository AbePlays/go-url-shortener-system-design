package api

import (
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/AbePlays/go-url-shortener-system-design/utils"
)

var templatesDir = func() string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	for {
		candidate := filepath.Join(dir, "web", "templates")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			log.Fatal("could not find web/templates directory")
		}
		dir = parent
	}
}()

var indexTpl = template.Must(template.ParseFiles(filepath.Join(templatesDir, "index.html")))
var linksTpl = template.Must(template.ParseFiles(filepath.Join(templatesDir, "links.html")))
var aboutTpl = template.Must(template.ParseFiles(filepath.Join(templatesDir, "about.html")))

func (handler *Handler) ShortenFormHandler(w http.ResponseWriter, req *http.Request) {
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

func (handler *Handler) LinksWebHandler(w http.ResponseWriter, req *http.Request) {
	links, err := handler.store.ListUrls(req.Context())
	if err != nil {
		slog.Error("failed to list urls", "error", err)
	}

	linksTpl.Execute(w, LinksPageData{Links: links})
}

func (handler *Handler) IndexWebHandler(w http.ResponseWriter, req *http.Request) {
	indexTpl.Execute(w, nil)
}

func (handler *Handler) AboutWebHandler(w http.ResponseWriter, req *http.Request) {
	aboutTpl.Execute(w, nil)
}
