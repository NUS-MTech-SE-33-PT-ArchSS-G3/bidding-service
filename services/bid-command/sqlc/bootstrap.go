package sqlc

import (
	"context"
	"database/sql"
	_ "embed"
)

var (
	//go:embed schema/001_init.sql
	ddl string
)

func EnsureSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, ddl)
	return err
}
