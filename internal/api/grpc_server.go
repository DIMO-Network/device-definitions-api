package api

import (
	"github.com/rs/zerolog"
	"net"

	"github.com/DIMO-Network/device-definitions-api/internal/api/common"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	pkggrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
)

func StartGrpcServer(logger zerolog.Logger, s *config.Settings, m mediator.Mediator) {
	lis, err := net.Listen("tcp", ":"+s.GRPCPort)
	if err != nil {
		logger.Fatal().Msgf("Failed to listen on port %v: %v", s.GRPCPort, err)
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

	logger.Info().Str("port", s.GRPCPort).Msgf("started grpc server on port: %v", s.GRPCPort)

	if err := server.Serve(lis); err != nil {
		logger.Fatal().Msgf("Failed to serve over port %v: %v", s.GRPCPort, err)
	}
}
