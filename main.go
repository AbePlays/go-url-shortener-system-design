package main

import (
	"log"
	"net/http"

	"github.com/AbePlays/go-url-shortener-system-design/internal/api"
	"github.com/AbePlays/go-url-shortener-system-design/internal/store"
)

func main() {
	urlStore := store.New()
	handler := api.New(urlStore)

	http.HandleFunc("POST /api/shorten", handler.ShortenHandler)
	http.HandleFunc("GET /{code}", handler.RedirectHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
