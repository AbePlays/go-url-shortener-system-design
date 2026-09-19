package main

import (
	"log"
	"net/http"
	"os"

	"github.com/AbePlays/go-url-shortener-system-design/internal/api"
	"github.com/AbePlays/go-url-shortener-system-design/internal/store"
)

func main() {
	port, baseUrl := getEnvKeys()

	urlStore := store.New()
	handler := api.New(urlStore, baseUrl)

	http.Handle("POST /api/shorten", api.LoggingMiddleware(http.HandlerFunc(handler.ShortenHandler)))
	http.Handle("GET /{code}", api.LoggingMiddleware(http.HandlerFunc(handler.RedirectHandler)))

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getEnvKeys() (string, string) {
	port := os.Getenv("PORT")
	baseUrl := os.Getenv("BASE_URL")
	if port == "" {
		log.Fatal("PORT environment variable not set")
	}
	if baseUrl == "" {
		log.Fatal("BASE_URL environment variable not set")
	}

	return port, baseUrl
}
