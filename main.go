package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/deshiwabudilaksana/fube-go/config"
	"github.com/deshiwabudilaksana/fube-go/database"
	"github.com/deshiwabudilaksana/fube-go/router"
)

func main() {
	// Load configuration
	cfg := config.Load()

	database.ConnectDB(cfg.DatabaseURL)

	r := router.NewRouter()

	fmt.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
