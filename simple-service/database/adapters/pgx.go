package adapters

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type Pgx func(ctx context.Context, connString string) (*pgx.Conn, error)

func (f Pgx) Connect(ctx context.Context, connString string) (*pgx.Conn, error) {
	return f(ctx, connString)
}
