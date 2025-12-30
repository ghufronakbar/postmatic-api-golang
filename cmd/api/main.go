// cmd/api/main.go
package main

import (
	"log"
	"net/http"
	"postmatic-api/config"
	handler "postmatic-api/internal/http"
	"postmatic-api/internal/http/middleware"

	_ "time/tzdata"

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
)

func main() {
	// 1. Load Config
	cfg := config.Load()

	// 2. Init DB
	db, err := config.ConnectDB(cfg.DATABASE_URL)
	if err != nil {
		log.Fatal("Cannot connect to DB:", err)
	}
	defer db.Close()

	// 3. Setup Router
	r := chi.NewRouter()
	r.Use(chiMw.RequestID)
	r.Use(middleware.RequestLogger)
	r.Use(chiMw.Recoverer)

	// 4. Register All Modules
	// Kita lempar DB ke fungsi init modules
	r.Mount("/api", handler.NewRouter(db))

	// 5. Start Server
	log.Println("Server running on port", cfg.PORT)
	if err := http.ListenAndServe(":"+cfg.PORT, r); err != nil {
		log.Fatal(err)
	}
}
