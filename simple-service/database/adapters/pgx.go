package adapters

import (
	"context"

	"github.com/addme96/simple-go-service/simple-service/database"
	"github.com/jackc/pgx/v4"
)

type Pgx struct{}

func (p Pgx) Connect(ctx context.Context, connString string) (database.PgxConn, error) {
	return pgx.Connect(ctx, connString)
}
