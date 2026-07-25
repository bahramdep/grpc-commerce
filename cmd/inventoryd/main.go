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

	inventoryv1 "github.com/bahramdep/grpc-commerce/gen/go/commerce/inventory/v1"
	"github.com/bahramdep/grpc-commerce/internal/inventory"
	"github.com/bahramdep/grpc-commerce/internal/inventory/grpcserver"
	"github.com/bahramdep/grpc-commerce/internal/inventory/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	defaultAddress  = ":50052"
	shutdownTimeout = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error(
			"inventory service failed",
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
		syscall.SIGTERM,
	)
	defer stop()

	address := os.Getenv("INVENTORYD_GRPC_ADDRESS")
	if address == "" {
		address = defaultAddress
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf(
			"failed to listen on %s: %w",
			address,
			err,
		)
	}
	defer listener.Close()

	repository := memory.NewRepository(map[string]int32{
		"product-1": 100,
		"product-2": 100,
	})

	inventoryService := inventory.NewService(repository)
	inventoryServer := grpcserver.New(inventoryService)

	grpcServer := grpc.NewServer()

	inventoryv1.RegisterInventoryServiceServer(
		grpcServer,
		inventoryServer,
	)
	reflection.Register(grpcServer)
	serveErrors := make(chan error, 1)

	go func() {
		serveErrors <- grpcServer.Serve(listener)
	}()

	slog.Info(
		"inventory gRPC server started",
		"address",
		address,
	)

	select {
	case err := <-serveErrors:
		if err != nil {
			return fmt.Errorf(
				"failed to serve inventory: %w",
				err,
			)
		}
		return nil

	case <-ctx.Done():
		slog.Info("shutting down inventory gRPC server")
		stopGracefully(grpcServer, shutdownTimeout)
		return nil
	}
}

func stopGracefully(
	server *grpc.Server,
	timeout time.Duration,
) {
	stopped := make(chan struct{})

	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-stopped:
		slog.Info("inventory gRPC service stopped")

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
