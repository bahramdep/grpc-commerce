package grpcserver

import (
	"context"
	"errors"

	inventoryv1 "github.com/bahramdep/grpc-commerce/gen/go/commerce/inventory/v1"
	"github.com/bahramdep/grpc-commerce/internal/inventory"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type InventoryService interface {
	Reserve(ctx context.Context, command inventory.ReserveCommand) (inventory.Reservation, error)
	Release(ctx context.Context, command inventory.ReleaseCommand) (inventory.Reservation, error)
}

type Server struct {
	inventoryv1.UnimplementedInventoryServiceServer
	service InventoryService
}

var _ inventoryv1.InventoryServiceServer = (*Server)(nil)

func New(service InventoryService) *Server {
	return &Server{
		service: service,
	}
}

func (s *Server) ReserveInventory(ctx context.Context, request *inventoryv1.ReserveInventoryRequest) (*inventoryv1.ReserveInventoryResponse, error) {
	items := make([]inventory.Item, len(request.GetItems()))
	for index, item := range request.GetItems() {
		items[index] = inventory.Item{
			ProductID: item.GetProductId(),
			Quantity:  item.GetQuantity(),
		}
	}

	command := inventory.ReserveCommand{IdempotencyKey: request.IdempotencyKey,
		OrderID: request.OrderId,
		Items:   items,
	}

	reservation, err := s.service.Reserve(ctx, command)
	if err != nil {
		return nil, toStatusError(err)
	}
	return &inventoryv1.ReserveInventoryResponse{Reservation: toProtoReservation(reservation)}, nil
}

func (s *Server) ReleaseInventory(
	ctx context.Context,
	request *inventoryv1.ReleaseInventoryRequest,
) (*inventoryv1.ReleaseInventoryResponse, error) {
	command := inventory.ReleaseCommand{
		IdempotencyKey: request.GetIdempotencyKey(),
		ReservationID:  request.GetReservationId()}

	reservation, err := s.service.Release(ctx, command)
	if err != nil {
		return nil, toStatusError(err)
	}
	return &inventoryv1.ReleaseInventoryResponse{Reservation: toProtoReservation(reservation)}, nil
}

func toProtoReservation(
	reservation inventory.Reservation,
) *inventoryv1.Reservation {
	items := make(
		[]*inventoryv1.InventoryItem,
		len(reservation.Items),
	)

	for index, item := range reservation.Items {
		items[index] = &inventoryv1.InventoryItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	return &inventoryv1.Reservation{
		Id:        reservation.ID,
		OrderId:   reservation.OrderID,
		Items:     items,
		Status:    toProtoStatus(reservation.Status),
		CreatedAt: timestamppb.New(reservation.CreatedAt),
	}
}

func toProtoStatus(
	inventoryStatus inventory.Status,
) inventoryv1.ReservationStatus {
	switch inventoryStatus {
	case inventory.StatusReserved:
		return inventoryv1.ReservationStatus_RESERVATION_STATUS_RESERVED
	case inventory.StatusReleased:
		return inventoryv1.ReservationStatus_RESERVATION_STATUS_RELEASED
	default:
		return inventoryv1.ReservationStatus_RESERVATION_STATUS_UNSPECIFIED
	}
}

func toStatusError(err error) error {
	switch {
	case errors.Is(err, inventory.ErrInvalidReserveInventory),
		errors.Is(err, inventory.ErrInvalidReleaseInventory):
		return status.Error(
			codes.InvalidArgument,
			"invalid inventory input",
		)

	case errors.Is(err, inventory.ErrIdempotencyKeyConflict):
		return status.Error(
			codes.AlreadyExists,
			"idempotency key conflict",
		)

	case errors.Is(err, inventory.ErrInsufficientStock):
		return status.Error(
			codes.FailedPrecondition,
			"insufficient stock",
		)

	case errors.Is(err, inventory.ErrReservationNotFound):
		return status.Error(
			codes.NotFound,
			"reservation not found",
		)

	case errors.Is(err, context.Canceled):
		return status.Error(
			codes.Canceled,
			"request canceled",
		)

	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(
			codes.DeadlineExceeded,
			"request deadline exceeded",
		)

	default:
		return status.Error(
			codes.Internal,
			"internal server error",
		)
	}
}
