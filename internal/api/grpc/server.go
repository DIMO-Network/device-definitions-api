package grpc

import (
	"log"
	"net"

	"github.com/DIMO-Network/device-definitions-api/internal/api/common"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	pkggrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
)

func StartGrpcServer(s *config.Settings, m mediator.Mediator) {
	lis, err := net.Listen("tcp", ":"+s.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %v: %v", s.GRPCPort, err)
	}

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(common.GrpcConfig),
	}

	service := NewGrpcService(m)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(opts...),
		)),
	)

	pkggrpc.RegisterDeviceDefinitionServiceServer(server, service)

	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve over port %v: %v", s.GRPCPort, err)
	}
}
