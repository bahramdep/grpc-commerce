package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	orderv1 "github.com/bahramdep/grpc-commerce/gen/go/commerce/order/v1"
	"github.com/bahramdep/grpc-commerce/internal/order"
	"github.com/bahramdep/grpc-commerce/internal/order/grpcserver"
	"github.com/bahramdep/grpc-commerce/internal/order/inventorygrpc"
	"github.com/bahramdep/grpc-commerce/internal/order/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	defaultAddress         = ":50051"
	defaultInventoryTarget = "localhost:50052"
	shutdownTimeout        = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error(
			"order service failed",
			"error",
			err,
		)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM)
	defer stop()

	address := os.Getenv("ORDERD_GRPC_ADDRESS")
	if address == "" {
		address = defaultAddress
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}
	defer lis.Close()

	inventoryTarget := os.Getenv("INVENTORY_GRPC_TARGET")
	if inventoryTarget == "" {
		inventoryTarget = defaultInventoryTarget
	}

	inventoryConnection, err := grpc.NewClient(
		inventoryTarget,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)
	if err != nil {
		return fmt.Errorf(
			"create Inventory gRPC client for %s: %w",
			inventoryTarget,
			err,
		)
	}
	defer inventoryConnection.Close()

	inventoryClient := inventorygrpc.New(inventoryConnection)

	repository := memory.NewRepository()
	orderService := order.NewService(repository, inventoryClient)
	orderServer := grpcserver.New(orderService)

	grpcServer := grpc.NewServer()

	orderv1.RegisterOrderServiceServer(
		grpcServer,
		orderServer)
	reflection.Register(grpcServer)

	serveErrors := make(chan error, 1)
	go func() {
		serveErrors <- grpcServer.Serve(lis)
	}()

	slog.Info(
		"order gRPC server started",
		"address", address)

	select {
	case err := <-serveErrors:
		if err != nil {
			return fmt.Errorf("failed to serve order: %w", err)
		}
		return nil
	case <-ctx.Done():
		slog.Info("shutting down order gRPC server")
		stopGracefully(grpcServer, shutdownTimeout)
		return nil
	}
}

func stopGracefully(server *grpc.Server, timeout time.Duration) {
	stopped := make(chan struct{})

	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-stopped:
		slog.Info("order gRPC service stopped")

	case <-timer.C:
		slog.Warn(
			"graceful shutdown timed out; forcing stop",
			"timeout",
			timeout,
		)

		server.Stop()
		<-stopped
	}
}
