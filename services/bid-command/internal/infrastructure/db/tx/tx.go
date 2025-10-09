package tx

import (
	"context"
	"database/sql"
	"kei-services/services/bid-command/internal/application"
)

var _ application.ITxManager = (*TxManager)(nil)

type TxManager struct{ DB *sql.DB }

func NewTxManager(db *sql.DB) TxManager {
	return TxManager{DB: db}
}

type txKey struct{}

// WithinTx executes the given function within a db transaction
func (m TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		// context already has a transaction, so just call the function
		return fn(ctx)
	}

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// inject tx into context
	ctxWithTx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(ctxWithTx); err != nil {
		// if error, rollback
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// FromCtx extracts the *sql.Tx from context
func FromCtx(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}
