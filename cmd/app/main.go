package main

import (
	"database/sql"
	"log"
	"net/http"

	"energosbyt/internal/config"
	"energosbyt/internal/handler"
	"energosbyt/internal/repository"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("clickhouse", cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer db.Close()

	repo := repository.NewClickHouseRepo(db)
	h := handler.NewHandler(repo)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	log.Printf("Starting server on port %s...", cfg.AppPort)
	if err = http.ListenAndServe(":"+cfg.AppPort, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
