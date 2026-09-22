package config

import (
	"log"
	"os"
)

type Config struct {
	BaseUrl     string
	DatabaseUrl string
	Port        string
	RedisUrl    string
}

func requireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable not set", key)
	}

	return value
}

func GetConfig() Config {
	baseUrl := requireEnv("BASE_URL")
	databaseUrl := requireEnv("DATABASE_URL")
	port := requireEnv("PORT")
	redisUrl := requireEnv("REDIS_URL")

	config := Config{
		BaseUrl:     baseUrl,
		DatabaseUrl: databaseUrl,
		Port:        port,
		RedisUrl:    redisUrl,
	}

	return config
}
