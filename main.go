package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AbePlays/go-url-shortener-system-design/internal/api"
	"github.com/AbePlays/go-url-shortener-system-design/internal/cache"
	"github.com/AbePlays/go-url-shortener-system-design/internal/config"
	"github.com/AbePlays/go-url-shortener-system-design/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
)

func main() {
	config := config.GetConfig()

	db, err := sql.Open("pgx", config.DatabaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	opts, err := redis.ParseURL(config.RedisUrl)
	if err != nil {
		log.Fatal(err)
	}

	redisClient := redis.NewClient(opts)
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	urlCache := cache.New(redisClient)
	urlStore := store.New(db, urlCache)
	handler := api.New(urlStore, config.BaseUrl)

	mux := http.NewServeMux()
	// Frontend Routes
	mux.HandleFunc("GET /{$}", handler.IndexWebHandler)
	mux.HandleFunc("GET /about", handler.AboutWebHandler)
	mux.HandleFunc("GET /links", handler.LinksWebHandler)

	// BFF Routes
	mux.HandleFunc("POST /shorten", handler.ShortenFormHandler)

	// Backend Routes
	mux.Handle("POST /api/shorten", api.LoggingMiddleware(http.HandlerFunc(handler.ShortenHandler)))
	mux.Handle("GET /{code}", api.LoggingMiddleware(http.HandlerFunc(handler.RedirectHandler)))

	server := &http.Server{
		Addr:    ":" + config.Port,
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
