package order

import (
	"context"
	"errors"
)

var ErrIdempotencyKeyConflict = errors.New("idempotency key conflict")

type Repository interface {
	Create(ctx context.Context, idempotencyKey string, candidate Order) (Order, error)
}
