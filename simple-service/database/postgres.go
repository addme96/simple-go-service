package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

type DB struct{}

func NewDB() *DB {
	return &DB{}
}

func (p *DB) GetConn() *pgx.Conn {
	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		os.Getenv("DB_ENDPOINT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
	)
	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	return conn
}
