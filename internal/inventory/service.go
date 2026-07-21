package inventory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidReserveInventory = errors.New(
		"invalid reserve inventory input",
	)
	ErrInvalidReleaseInventory = errors.New(
		"invalid release inventory input",
	)
)

type ReserveCommand struct {
	IdempotencyKey string
	OrderID        string
	Items          []Item
}

type ReleaseCommand struct {
	IdempotencyKey string
	Reservation_id string
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
		now:        time.Now,
	}
}

func (s *Service) Reserve(ctx context.Context, command ReserveCommand) (Reservation, error) {
	if err := validateReserve(command); err != nil {
		return Reservation{}, err
	}

	condidate := Reservation{
		OrderID:   command.OrderID,
		Items:     command.Items,
		Status:    StatusReserved,
		CreatedAt: s.now().UTC(),
	}

	return s.repository.Reserve(ctx, command.IdempotencyKey, condidate)
}

func (s *Service) Releas(
	ctx context.Context,
	command ReleaseCommand,
) (Reservation, error) {
	if err := validateRelease(command); err != nil {
		return Reservation{}, err
	}

	return s.repository.Release(ctx, command.IdempotencyKey, command.Reservation_id)
}

func validateReserve(command ReserveCommand) error {
	if strings.TrimSpace(command.IdempotencyKey) == "" {
		return fmt.Errorf(
			"%w: idempotency key is required",
			ErrInvalidReserveInventory,
		)
	}
	if strings.TrimSpace(command.OrderID) == "" {
		return fmt.Errorf(
			"%w: order id is required",
			ErrInvalidReserveInventory,
		)
	}

	if len(command.Items) == 0 {
		return fmt.Errorf(
			"%w: items are required",
			ErrInvalidReserveInventory,
		)
	}

	// hash map to check duplication items;
	productIDs := make(map[string]any, len(command.Items))
	for index, item := range command.Items {
		productID := strings.TrimSpace(item.ProductID)

		if productID == "" {
			return fmt.Errorf(
				"%w: item %d product id is required",
				ErrInvalidReserveInventory,
				index,
			)
		}

		if item.Quantity <= 0 {
			return fmt.Errorf(
				"%w: item %d quantity must be greater than zero",
				ErrInvalidReserveInventory,
				index,
			)
		}

		if _, found := productIDs[productID]; found {
			return fmt.Errorf(
				"%w: item %d duplicates product %q",
				ErrInvalidReserveInventory,
				index,
				productID,
			)
		}
		productIDs[productID] = struct{}{}
	}

	return nil
}

func validateRelease(command ReleaseCommand) error {
	if strings.TrimSpace(command.IdempotencyKey) == "" {
		return fmt.Errorf(
			"%w: idempotency key is required",
			ErrInvalidReleaseInventory,
		)
	}

	if strings.TrimSpace(command.Reservation_id) == "" {
		return fmt.Errorf(
			"%w: reservation ID is required",
			ErrInvalidReleaseInventory,
		)
	}
	return nil
}
