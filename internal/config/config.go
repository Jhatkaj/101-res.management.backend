package config

import "os"

type Config struct{ DatabaseURL, HTTPAddr, CORSOrigin, AdminUsername, AdminPassword, OwnerUsername, OwnerPassword string }

func Load() Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	origin := os.Getenv("CORS_ORIGIN")
	if origin == "" {
		origin = "*"
	}
	adminUsername, adminPassword := os.Getenv("ADMIN_USERNAME"), os.Getenv("ADMIN_PASSWORD")
	if adminUsername == "" {
		adminUsername = "admin"
	}
	if adminPassword == "" {
		adminPassword = "admin1234"
	}
	ownerUsername, ownerPassword := os.Getenv("OWNER_USERNAME"), os.Getenv("OWNER_PASSWORD")
	if ownerUsername == "" {
		ownerUsername = "owner"
	}
	if ownerPassword == "" {
		ownerPassword = "owner1234"
	}
	return Config{DatabaseURL: os.Getenv("DATABASE_URL"), HTTPAddr: addr, CORSOrigin: origin, AdminUsername: adminUsername, AdminPassword: adminPassword, OwnerUsername: ownerUsername, OwnerPassword: ownerPassword}
}
