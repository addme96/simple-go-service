package adapters

import (
	"context"

	"github.com/addme96/simple-go-service/simple-service/database"
	"github.com/jackc/pgx/v4"
)

type Pgx func(ctx context.Context, connString string) (*pgx.Conn, error)

func (f Pgx) Connect(ctx context.Context, connString string) (database.PgxConn, error) {
	return f(ctx, connString)
}
