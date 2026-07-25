package order

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const inventoryTimeout = 2 * time.Second

var ErrInvalidCreateOrder = errors.New("invalid create order input")
var ErrInvalidOrderState = errors.New("invalid order state")

type CreateCommand struct {
	IdempotencyKey string
	CustomerID     string
	Items          []Item
}

func validateCreate(command CreateCommand) error {
	if strings.TrimSpace(command.IdempotencyKey) == "" {
		return fmt.Errorf(
			"%w: idempotency key is required",
			ErrInvalidCreateOrder,
		)
	}

	if strings.TrimSpace(command.CustomerID) == "" {
		return fmt.Errorf(
			"%w: customer ID is required",
			ErrInvalidCreateOrder,
		)
	}
	if len(command.Items) == 0 {
		return fmt.Errorf(
			"%w: items are required",
			ErrInvalidCreateOrder,
		)
	}

	for index, item := range command.Items {
		if strings.TrimSpace(item.ProductID) == "" {
			return fmt.Errorf(
				"%w: item %d product ID is required",
				ErrInvalidCreateOrder,
				index,
			)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf(
				"%w: item %d quantity must be greater than zero",
				ErrInvalidCreateOrder,
				index,
			)
		}
	}

	return nil
}

type Service struct {
	repository Repository
	inventory  Inventory
	now        func() time.Time
}

func NewService(
	repository Repository,
	inventory Inventory,
) *Service {
	return &Service{
		repository: repository,
		inventory:  inventory,
		now:        time.Now,
	}
}

func (s *Service) Create(ctx context.Context, command CreateCommand) (Order, error) {
	if err := validateCreate(command); err != nil {
		return Order{}, err
	}
	items := make([]Item, len(command.Items))

	for index, item := range command.Items {
		items[index] = Item{
			ProductID: strings.TrimSpace(item.ProductID),
			Quantity:  item.Quantity,
		}
	}
	candidate := Order{
		CustomerID: command.CustomerID,
		Items:      items,
		Status:     StatusPending,
		CreatedAt:  s.now().UTC(),
	}

	created, err := s.repository.Create(ctx, command.IdempotencyKey, candidate)
	if err != nil {
		return Order{}, err
	}

	if created.Status == StatusConfirmed {
		return created, nil
	}

	if created.Status != StatusPending {
		return Order{}, fmt.Errorf(
			"%w: order %q has status %d",
			ErrInvalidOrderState,
			created.ID,
			created.Status,
		)
	}
	reservationContext, cancel := context.WithTimeout(ctx, inventoryTimeout)
	defer cancel()

	reservation, err := s.inventory.Reserve(
		reservationContext,
		ReserverInventoryCommand{
			IdempotencyKey: fmt.Sprintf(
				"order:%s:reserve",
				created.ID,
			),
			OrderID: created.ID,
			Items:   created.Items,
		},
	)
	if err != nil {
		return Order{}, err
	}
	return s.repository.ConfirmInventory(ctx, created.ID, reservation.ID)
}
