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
	"github.com/jackc/pgx/v4"
)

const (
	envDBEndpoint = "DB_ENDPOINT"
	envDBUsername = "DB_USERNAME"
	envDBName     = "DB_NAME"
	envDBPassword = "DB_PASSWORD"
)

func main() {
	db := database.NewDB(adapters.Pgx(pgx.Connect), buildConnectionStringFromEnv())
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

func buildConnectionStringFromEnv() string {
	dbEndpoint, ok := os.LookupEnv(envDBEndpoint)
	if !ok {
		panic(fmt.Sprintf("required %s is not set", envDBEndpoint))
	}
	dbUsername, ok := os.LookupEnv(envDBUsername)
	if !ok {
		panic(fmt.Sprintf("required %s is not set", envDBUsername))
	}
	dbName, ok := os.LookupEnv(envDBName)
	if !ok {
		panic(fmt.Sprintf("required %s is not set", envDBName))
	}
	dbPassword, ok := os.LookupEnv(envDBPassword)
	if !ok {
		panic(fmt.Sprintf("required %s is not set", envDBPassword))
	}
	return fmt.Sprintf("postgres://%s:%s@%s/%s", dbEndpoint, dbUsername, dbName, dbPassword)
}
