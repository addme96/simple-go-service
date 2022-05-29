package main

import (
	"context"
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
	db := database.NewDB(adapters.Pgx(pgx.Connect), getConnectionString())
	ctx := context.Background()

	if err := database.Seed(ctx, db); err != nil {
		panic(err)
	}
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

func getConnectionString() string {
	env, err := readAllEnvVars(envDBEndpoint, envDBUsername, envDBName, envDBPassword)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("postgres://%s:%s@%s/%s",
		env[envDBUsername],
		env[envDBPassword],
		env[envDBEndpoint],
		env[envDBName],
	)
}

func readAllEnvVars(keys ...string) (map[string]string, error) {
	env := make(map[string]string, len(keys))
	for _, name := range keys {
		value, ok := os.LookupEnv(name)
		if !ok {
			return nil, fmt.Errorf("required %s is not set", name)
		}
		env[name] = value
	}
	return env, nil
}
