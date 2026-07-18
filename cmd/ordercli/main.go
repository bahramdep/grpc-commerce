package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	orderv1 "github.com/bahramdep/grpc-commerce/gen/go/commerce/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	defaultTarget = "localhost:50051"
	callTimeout   = 3 * time.Second
)

func main() {
	if err := run(); err != nil {
		if grpcStatus, ok := status.FromError(err); ok {
			slog.Error(
				"CreateOrder RPC failed",
				"code",
				grpcStatus.Code(),
				"message",
				grpcStatus.Message(),
			)
		} else {
			slog.Error(
				"order client failed",
				"error",
				err,
			)
		}

		os.Exit(1)
	}
}

func run() error {
	target := os.Getenv("ORDER_GRPC_TARGET")
	if target == "" {
		target = defaultTarget
	}

	connection, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)
	if err != nil {
		return fmt.Errorf(
			"create gRPC client for %s: %w",
			target,
			err,
		)
	}
	defer connection.Close()

	client := orderv1.NewOrderServiceClient(connection)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		callTimeout,
	)
	defer cancel()

	response, err := client.CreateOrder(
		ctx,
		&orderv1.CreateOrderRequest{
			IdempotencyKey: "demo-request-1",
			CustomerId:     "customer-1",
			Items: []*orderv1.OrderItem{
				{
					ProductId: "product-1",
					Quantity:  2,
				},
				{
					ProductId: "product-2",
					Quantity:  1,
				},
			},
		},
	)
	if err != nil {
		return err
	}

	output, err := protojson.MarshalOptions{
		Multiline: true,
		Indent:    "  ",
	}.Marshal(response.GetOrder())
	if err != nil {
		return fmt.Errorf(
			"encode CreateOrder response: %w",
			err,
		)
	}

	fmt.Println(string(output))

	return nil
}
