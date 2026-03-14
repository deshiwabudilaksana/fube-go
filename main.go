package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/deshiwabudilaksana/fube-go/config"
	"github.com/deshiwabudilaksana/fube-go/database"
	"github.com/deshiwabudilaksana/fube-go/db"
	"github.com/deshiwabudilaksana/fube-go/router"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	database.ConnectDB(cfg.DatabaseURL)

	// MongoDB local connection check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.GetConnection(ctx, cfg)
	if err != nil {
		log.Printf("Warning: Could not connect to local MongoDB: %v", err)
	} else {
		log.Println("MongoDB Connected Locally")
	}

	r := router.NewRouter()

	log.Printf("Server starting on %s", cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
