package application

import "context"

type ITxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
