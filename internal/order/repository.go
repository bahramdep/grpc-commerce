package order

import "context"

type Repository interface {
	Create(ctx context.Context, candidate Order) (Order, error)
}
