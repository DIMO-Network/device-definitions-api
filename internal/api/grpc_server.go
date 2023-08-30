package api

import (
	"net"

	"github.com/DIMO-Network/device-definitions-api/internal/api/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	pkggrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func StartGrpcServer(logger zerolog.Logger, s *config.Settings, m mediator.Mediator) {
	lis, err := net.Listen("tcp", ":"+s.GRPCPort)
	if err != nil {
		logger.Fatal().Msgf("Failed to listen on port %v: %v", s.GRPCPort, err)
	}

	deviceDefinitionService := NewGrpcService(m, &logger)
	recallsService := NewGrpcRecallsService(m, &logger)
	reviewsService := NewGrpcReviewsService(m, &logger)
	integrationService := NewGrpcIntegrationService(m, &logger)
	decodeService := NewGrpcVinDecoderService(m, &logger)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			metrics.ValidationMiddleware(),
			recoveryMiddleware(),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
		)),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	pkggrpc.RegisterDeviceDefinitionServiceServer(server, deviceDefinitionService)
	pkggrpc.RegisterRecallsServiceServer(server, recallsService)
	pkggrpc.RegisterReviewsServiceServer(server, reviewsService)
	pkggrpc.RegisterIntegrationServiceServer(server, integrationService)
	pkggrpc.RegisterVinDecoderServiceServer(server, decodeService)

	// Register reflection service on gRPC server
	reflection.Register(server)

	// Register the gRPC Prometheus metrics collector.
	grpc_prometheus.Register(server)

	logger.Info().Str("port", s.GRPCPort).Msgf("started grpc server on port: %v", s.GRPCPort)

	if err := server.Serve(lis); err != nil {
		logger.Fatal().Msgf("Failed to serve over port %v: %v", s.GRPCPort, err)
	}
}

func recoveryMiddleware() grpc.UnaryServerInterceptor {
	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(common.GrpcConfig),
	}
	return grpc_recovery.UnaryServerInterceptor(recoveryOpts...)
}
