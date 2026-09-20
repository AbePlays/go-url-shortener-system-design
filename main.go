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
	"github.com/AbePlays/go-url-shortener-system-design/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	config := getConfig()

	db, err := sql.Open("pgx", config.DatabaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	urlStore := store.New(db)
	handler := api.New(urlStore, config.BaseUrl)

	mux := http.NewServeMux()
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

type Config struct {
	Port        string
	BaseUrl     string
	DatabaseUrl string
}

func requireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable not set", key)
	}
	return value
}

func getConfig() Config {
	port := requireEnv("PORT")
	baseUrl := requireEnv("BASE_URL")
	databaseUrl := requireEnv("DATABASE_URL")

	config := Config{
		Port:        port,
		BaseUrl:     baseUrl,
		DatabaseUrl: databaseUrl,
	}

	return config
}
