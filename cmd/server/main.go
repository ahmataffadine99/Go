package main

import (
	"log"
	"net/http"

	"ecommerce-cli/internal/config"
	"ecommerce-cli/internal/database"
	"ecommerce-cli/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.InitDB(cfg.DBDriver, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db, "internal/database/schema.sql"); err != nil {
		log.Printf("migration warning: %v", err)
	}

	router := server.NewRouter(db, cfg)

	log.Printf("Starting HTTP backend server on port :%s ...", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, router); err != nil {
		log.Fatalf("server failure: %v", err)
	}
}
