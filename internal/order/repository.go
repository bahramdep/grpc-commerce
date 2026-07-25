package order

import (
	"context"
	"errors"
)

var (
	ErrIdempotencyKeyConflict = errors.New(
		"idempotency key conflict",
	)
	ErrOrderNotFound = errors.New(
		"order not found",
	)
	ErrInventoryReservationConflict = errors.New(
		"inventory reservation conflict",
	)
)

type Repository interface {
	Create(ctx context.Context, idempotencyKey string, candidate Order) (Order, error)
	ConfirmInventory(ctx context.Context, orderID string, reservationID string) (Order, error)
}
