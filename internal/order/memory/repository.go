package memory

import (
	"context"
	"strconv"
	"sync"

	"github.com/bahramdep/grpc-commerce/internal/order"
)

type Repository struct {
	mu     sync.Mutex
	nextID uint64
	orders map[string]order.Order
}

var _ order.Repository = (*Repository)(nil)

func NewRepository() *Repository {
	return &Repository{
		orders: make(map[string]order.Order),
	}
}

func (r *Repository) Create(ctx context.Context, candidate order.Order) (order.Order, error) {
	if err := ctx.Err(); err != nil {
		return order.Order{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	candidate.ID = strconv.FormatUint(r.nextID, 10)

	stored := cloneOrder(candidate)
	r.orders[stored.ID] = stored

	return cloneOrder(stored), nil
}

func cloneOrder(source order.Order) order.Order {
	cloned := source
	cloned.Items = append([]order.Item(nil), source.Items...)
	return cloned
}
