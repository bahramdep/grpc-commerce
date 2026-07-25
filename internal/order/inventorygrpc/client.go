package inventorygrpc

import (
	"context"
	"fmt"
	"strings"

	inventoryv1 "github.com/bahramdep/grpc-commerce/gen/go/commerce/inventory/v1"
	"github.com/bahramdep/grpc-commerce/internal/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	client inventoryv1.InventoryServiceClient
}

var _ order.Inventory = (*Client)(nil)

func New(connection grpc.ClientConnInterface) *Client {
	return &Client{
		client: inventoryv1.NewInventoryServiceClient(connection),
	}
}

func (c *Client) Reserve(
	ctx context.Context,
	command order.ReserveInventoryCommand,
) (order.InventoryReservation, error) {
	items := make(
		[]*inventoryv1.InventoryItem,
		len(command.Items),
	)

	for index, item := range command.Items {
		items[index] = &inventoryv1.InventoryItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	response, err := c.client.ReserveInventory(
		ctx,
		&inventoryv1.ReserveInventoryRequest{
			IdempotencyKey: command.IdempotencyKey,
			OrderId:        command.OrderID,
			Items:          items,
		},
	)
	if err != nil {
		return order.InventoryReservation{},
			toDomainError(err)
	}

	reservation := response.GetReservation()

	if reservation == nil {
		return order.InventoryReservation{}, fmt.Errorf(
			"%w: reservation is missing",
			order.ErrInvalidInventoryResponse,
		)
	}

	if strings.TrimSpace(reservation.GetId()) == "" {
		return order.InventoryReservation{}, fmt.Errorf(
			"%w: reservation ID is missing",
			order.ErrInvalidInventoryResponse,
		)
	}

	if reservation.GetOrderId() != command.OrderID {
		return order.InventoryReservation{}, fmt.Errorf(
			"%w: expected order %q, received %q",
			order.ErrInvalidInventoryResponse,
			command.OrderID,
			reservation.GetOrderId(),
		)
	}

	if reservation.GetStatus() !=
		inventoryv1.ReservationStatus_RESERVATION_STATUS_RESERVED {
		return order.InventoryReservation{}, fmt.Errorf(
			"%w: unexpected reservation status %s",
			order.ErrInvalidInventoryResponse,
			reservation.GetStatus(),
		)
	}

	return order.InventoryReservation{
		ID:      reservation.GetId(),
		OrderID: reservation.GetOrderId(),
	}, nil
}

func toDomainError(err error) error {
	switch status.Code(err) {
	case codes.FailedPrecondition:
		return fmt.Errorf(
			"%w: %v",
			order.ErrInsufficientInventory,
			err,
		)

	case codes.AlreadyExists:
		return fmt.Errorf(
			"%w: %v",
			order.ErrInventoryConflict,
			err,
		)

	case codes.Unavailable:
		return fmt.Errorf(
			"%w: %v",
			order.ErrInventoryUnavailable,
			err,
		)

	case codes.DeadlineExceeded:
		return fmt.Errorf(
			"%w: inventory request deadline exceeded",
			context.DeadlineExceeded,
		)

	case codes.Canceled:
		return fmt.Errorf(
			"%w: inventory request canceled",
			context.Canceled,
		)

	default:
		return fmt.Errorf(
			"%w: %v",
			order.ErrInventoryFailure,
			err,
		)
	}
}
