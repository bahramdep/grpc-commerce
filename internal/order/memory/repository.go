package memory

import (
	"context"
	"slices"
	"strconv"
	"sync"

	"github.com/bahramdep/grpc-commerce/internal/order"
)

type Repository struct {
	mu              sync.Mutex
	nextID          uint64
	orders          map[string]order.Order
	idempotencyKeys map[string]string
}

//var _ order.Repository = (*Repository)(nil)

func NewRepository() *Repository {
	return &Repository{
		orders:          make(map[string]order.Order),
		idempotencyKeys: make(map[string]string),
	}
}

func (r *Repository) Create(ctx context.Context, idempotencyKey string, candidate order.Order) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return order.Order{}, err
	}

	if existingID, found := r.idempotencyKeys[idempotencyKey]; found {
		existing := r.orders[existingID]

		sameRequest := existing.CustomerID == candidate.CustomerID &&
			slices.Equal(existing.Items, candidate.Items)
		if !sameRequest {
			return order.Order{}, order.ErrIdempotencyKeyConflict
		}
		return cloneOrder(existing), nil
	}

	r.nextID++
	candidate.ID = strconv.FormatUint(r.nextID, 10)

	stored := cloneOrder(candidate)
	r.orders[stored.ID] = stored
	r.idempotencyKeys[idempotencyKey] = stored.ID

	return cloneOrder(stored), nil
}

func (r *Repository) ConfirmInventory(ctx context.Context, orderID string, reservationID string) (order.Order, error) {
	if err := ctx.Err(); err != nil {
		return order.Order{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, found := r.orders[orderID]
	if !found {
		return order.Order{}, order.ErrOrderNotFound
	}

	if existing.Status == order.StatusConfirmed {
		if existing.InventoryReservationID != reservationID {
			return order.Order{}, order.ErrInventoryReservationConflict
		}
		return cloneOrder(existing), nil
	}

	existing.Status = order.StatusConfirmed
	existing.InventoryReservationID = reservationID

	stored := cloneOrder(existing)
	r.orders[orderID] = stored
	return cloneOrder(existing), nil
}

func cloneOrder(source order.Order) order.Order {
	cloned := source
	cloned.Items = slices.Clone(source.Items)
	return cloned
}
