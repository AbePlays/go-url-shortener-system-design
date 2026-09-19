package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AbePlays/go-url-shortener-system-design/internal/api"
	"github.com/AbePlays/go-url-shortener-system-design/internal/store"
)

func main() {
	port, baseUrl := getEnvKeys()

	urlStore := store.New()
	handler := api.New(urlStore, baseUrl)

	mux := http.NewServeMux()
	mux.Handle("POST /api/shorten", api.LoggingMiddleware(http.HandlerFunc(handler.ShortenHandler)))
	mux.Handle("GET /{code}", api.LoggingMiddleware(http.HandlerFunc(handler.RedirectHandler)))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("server shut down gracefully")
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
