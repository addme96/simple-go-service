//go:generate mockgen -destination=mocks/pgx.go -package mocks . Pgx
package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type Pgx interface {
	Connect(ctx context.Context, connString string) (PgxConn, error)
}

type PgxConn interface {
	Prepare(ctx context.Context, name, sql string) (sd *pgconn.StatementDescription, err error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Close(ctx context.Context) error
}

type DB struct {
	pgx         Pgx
	databaseURL string
}

func NewDB(pgx Pgx, databaseURL string) *DB {
	return &DB{pgx: pgx, databaseURL: databaseURL}
}

func (p *DB) GetConn(ctx context.Context) (PgxConn, error) {
	conn, err := p.pgx.Connect(ctx, p.databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return nil, err
	}
	return conn, nil
}
