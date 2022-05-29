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
	Connect(ctx context.Context, connString string) (*pgx.Conn, error)
}

//PgxConn allows using pgxmock in tests
type PgxConn interface {
	Begin(context.Context) (pgx.Tx, error)
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	Ping(context.Context) error
	Prepare(context.Context, string, string) (*pgconn.StatementDescription, error)
	Close(context.Context) error
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

func Seed(ctx context.Context, db *DB) error {
	conn, err := db.GetConn(ctx)
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS resources (
id INT GENERATED ALWAYS AS IDENTITY, 
name varchar
)`)
	return err
}
