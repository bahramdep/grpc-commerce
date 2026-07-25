package memory

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"sync"

	"github.com/bahramdep/grpc-commerce/internal/inventory"
)

type Repository struct {
	mu          sync.Mutex
	nextID      uint64
	stock       map[string]int32
	reservation map[string]inventory.Reservation
	reserveKeys map[string]string
	releaseKeys map[string]string
}

func NewRepository(initialStock map[string]int32) *Repository {
	stock := make(map[string]int32, len(initialStock))
	for productID, quantity := range initialStock {
		stock[productID] = quantity
	}

	return &Repository{
		stock:       stock,
		reservation: make(map[string]inventory.Reservation),
		reserveKeys: make(map[string]string),
		releaseKeys: make(map[string]string),
	}
}

func (r *Repository) Reserve(
	ctx context.Context,
	idempotencyKey string,
	candidate inventory.Reservation,
) (inventory.Reservation, error) {
	if err := ctx.Err(); err != nil {
		return inventory.Reservation{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if reservationID, found := r.reserveKeys[idempotencyKey]; found {
		existing := r.reservation[reservationID]
		sameRequest := existing.OrderID == candidate.OrderID &&
			slices.Equal(existing.Items, candidate.Items)
		if !sameRequest {
			return inventory.Reservation{},
				inventory.ErrIdempotencyKeyConflict
		}
		return cloneReservation(existing), nil
	}
	for _, item := range candidate.Items {
		available := r.stock[item.ProductID]

		if available < item.Quantity {
			return inventory.Reservation{}, fmt.Errorf(
				"%w: product %q requested %d, available %d",
				inventory.ErrInsufficientStock, item.ProductID, item.Quantity, available)
		}
	}

	for _, item := range candidate.Items {
		r.stock[item.ProductID] -= item.Quantity
	}
	r.nextID++
	candidate.ID = strconv.FormatUint(r.nextID, 10)

	stored := cloneReservation(candidate)
	r.reservation[stored.ID] = stored
	r.reserveKeys[idempotencyKey] = stored.ID

	return cloneReservation(stored), nil
}

func (r *Repository) Release(
	ctx context.Context,
	idempotencyKey string,
	reservationID string,
) (inventory.Reservation, error) {
	if err := ctx.Err(); err != nil {
		return inventory.Reservation{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if existingID, found := r.releaseKeys[idempotencyKey]; found {
		if existingID != reservationID {
			return inventory.Reservation{},
				inventory.ErrIdempotencyKeyConflict
		}

		return cloneReservation(r.reservation[existingID]), nil
	}

	reservation, found := r.reservation[reservationID]
	if !found {
		return inventory.Reservation{},
			inventory.ErrReservationNotFound
	}

	// A second release with a new idempotency key must not restore
	// the stock for a second time.
	if reservation.Status == inventory.StatusReleased {
		r.releaseKeys[idempotencyKey] = reservationID
		return cloneReservation(reservation), nil
	}

	for _, item := range reservation.Items {
		r.stock[item.ProductID] += item.Quantity
	}

	reservation.Status = inventory.StatusReleased

	stored := cloneReservation(reservation)
	r.reservation[reservationID] = stored
	r.releaseKeys[idempotencyKey] = reservationID

	return cloneReservation(stored), nil
}

func cloneReservation(
	source inventory.Reservation,
) inventory.Reservation {
	cloned := source
	cloned.Items = slices.Clone(source.Items)

	return cloned
}
