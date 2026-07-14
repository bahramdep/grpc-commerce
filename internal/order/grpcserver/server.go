package grpcserver

import (
	orderv1 "github.com/bahramdep/grpc-commerce/gen/go/commerce/order/v1"
)

type Server struct {
	orderv1.UnimplementedOrderServiceServer
}

var _ orderv1.OrderServiceServer = (*Server)(nil)

func New() *Server {
	return &Server{}
}
