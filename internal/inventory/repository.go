package inventory

import (
	"context"
	"errors"
)

var (
	ErrIdempotencyKeyConflict = errors.New(
		"idempotency key conflict",
	)
	ErrInsufficientStock = errors.New(
		"insufficient stock",
	)
	ErrReservationNotFound = errors.New(
		"reservation not found",
	)
)

type Repository interface {
	Reserve(
		ctx context.Context,
		idempotencyKey string,
		candidate Reservation,
	) (Reservation, error)

	Release(
		ctx context.Context,
		idempotencyKey string,
		reservationID string,
	) (Reservation, error)
}
