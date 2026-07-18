package order

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidCreateOrder = errors.New("invalid create order input")

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
	now        func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repository: repo,
		now:        time.Now,
	}
}

func (s *Service) Create(ctx context.Context, command CreateCommand) (Order, error) {
	if err := validateCreate(command); err != nil {
		return Order{}, err
	}

	candidate := Order{
		CustomerID: command.CustomerID,
		Items:      command.Items,
		Status:     StatusPending,
		CreatedAt:  s.now().UTC(),
	}

	return s.repository.Create(ctx, command.IdempotencyKey, candidate)
}
