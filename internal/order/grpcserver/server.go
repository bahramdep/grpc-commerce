package grpcserver

import (
	"context"
	"errors"

	orderv1 "github.com/bahramdep/grpc-commerce/gen/go/commerce/order/v1"
	"github.com/bahramdep/grpc-commerce/internal/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Creator interface {
	Create(ctx context.Context, command order.CreateCommand) (order.Order, error)
}

type Server struct {
	orderv1.UnimplementedOrderServiceServer

	creator Creator
}

var _ orderv1.OrderServiceServer = (*Server)(nil)

func New(creator Creator) *Server {
	return &Server{
		creator: creator,
	}
}

func (s *Server) CreateOrder(
	ctx context.Context,
	request *orderv1.CreateOrderRequest,
) (*orderv1.CreateOrderResponse, error) {
	items := make([]order.Item, len(request.GetItems()))

	for index, item := range request.GetItems() {
		items[index] = order.Item{
			ProductID: item.GetProductId(),
			Quantity:  item.GetQuantity(),
		}
	}

	command := order.CreateCommand{
		IdempotencyKey: request.GetIdempotencyKey(),
		CustomerID:     request.GetCustomerId(),
		Items:          items,
	}
	created, err := s.creator.Create(ctx, command)
	if err != nil {
		return nil, toStatusError(err)
	}

	return &orderv1.CreateOrderResponse{
		Order: toProtoOrder(created),
	}, nil
}

func toProtoOrder(order order.Order) *orderv1.Order {
	items := make([]*orderv1.OrderItem, len(order.Items))
	for index, item := range order.Items {
		items[index] = &orderv1.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	return &orderv1.Order{
		Id:         order.ID,
		CustomerId: order.CustomerID,
		Items:      items,
		Status:     toProtoStatus(order.Status),
		CreatedAt:  timestamppb.New(order.CreatedAt),
	}
}

func toProtoStatus(status order.Status) orderv1.OrderStatus {
	switch status {
	case order.StatusPending:
		return orderv1.OrderStatus_ORDER_STATUS_PENDING
	default:
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func toStatusError(err error) error {
	switch {
	case errors.Is(err, order.ErrInvalidCreateOrder):
		return status.Error(codes.InvalidArgument, "invalid create order input")
	case errors.Is(err, order.ErrIdempotencyKeyConflict):
		return status.Error(codes.AlreadyExists, "idempotency key conflict")
	case errors.Is(err, context.Canceled),
		errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.Canceled, "request canceled")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
