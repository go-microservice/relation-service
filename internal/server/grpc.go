package server

import (
	"time"

	"github.com/go-eagle/eagle/pkg/app"
	"github.com/go-eagle/eagle/pkg/transport/grpc"

	v1 "github.com/go-microservice/relation-service/api/relation/v1"
	"github.com/go-microservice/relation-service/internal/service"
)

// NewGRPCServer creates a gRPC server
func NewGRPCServer(cfg *app.ServerConfig, svc *service.RelationServiceServer) *grpc.Server {

	grpcServer := grpc.NewServer(
		grpc.Network("tcp"),
		grpc.Address(cfg.Addr),
		grpc.Timeout(3*time.Second),
	)

	// register biz service
	v1.RegisterRelationServiceServer(grpcServer, svc)

	return grpcServer
}
