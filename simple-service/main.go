package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/addme96/simple-go-service/simple-service/database"
	"github.com/addme96/simple-go-service/simple-service/database/adapters"
	"github.com/addme96/simple-go-service/simple-service/handlers"
	"github.com/addme96/simple-go-service/simple-service/repositories"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	db := database.NewDB(adapters.Pgx{}, fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		os.Getenv("DB_ENDPOINT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
	))
	resourceHandler := handlers.NewResource(repositories.NewResource(db))
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Route("/resources", func(r chi.Router) {
		r.Get("/", resourceHandler.List)
		r.Post("/", resourceHandler.Post)
		r.Route("/{resourceID}", func(r chi.Router) {
			r.Use(resourceHandler.GetCtx)
			r.Get("/", resourceHandler.Get)
			r.Put("/", resourceHandler.Put)
			r.Delete("/", resourceHandler.Delete)
		})
	})
	log.Println("Listening for requests at http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
