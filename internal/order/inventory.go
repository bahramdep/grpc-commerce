package order

import (
	"context"
	"errors"
)

var (
	ErrInsufficientInventory = errors.New(
		"insufficient inventory",
	)
	ErrInventoryUnavailable = errors.New(
		"inventory service unavailable",
	)
	ErrInventoryConflict = errors.New(
		"inventory idempotency conflict",
	)
	ErrInventoryFailure = errors.New(
		"inventory operation failed",
	)
	ErrInvalidInventoryResponse = errors.New(
		"invalid inventory response",
	)
)

type ReserveInventoryCommand struct {
	IdempotencyKey string
	OrderID        string
	Items          []Item
}

type InventoryReservation struct {
	ID      string
	OrderID string
}

type Inventory interface {
	Reserve(
		ctx context.Context,
		command ReserveInventoryCommand,
	) (InventoryReservation, error)
}
