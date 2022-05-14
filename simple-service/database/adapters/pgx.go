package adapters

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type Pgx struct{}

func NewPgx() *Pgx {
	return &Pgx{}
}

func (p Pgx) Connect(ctx context.Context, connString string) (*pgx.Conn, error) {
	return pgx.Connect(ctx, connString)
}
