package config

import "os"

type Config struct { DatabaseURL, HTTPAddr, CORSOrigin string }

func Load() Config {
	addr := os.Getenv("HTTP_ADDR"); if addr == "" { addr = ":8080" }
	origin := os.Getenv("CORS_ORIGIN"); if origin == "" { origin = "*" }
	return Config{DatabaseURL: os.Getenv("DATABASE_URL"), HTTPAddr: addr, CORSOrigin: origin}
}
