package api

import (
	"net"

	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/gateways"
	"github.com/DIMO-Network/shared/db"

	"github.com/DIMO-Network/device-definitions-api/internal/api/common"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/core/mediator"
	pkggrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func StartGrpcServer(logger zerolog.Logger, s *config.Settings, m mediator.Mediator, dbs func() *db.ReaderWriter, onChainDeviceDefs gateways.DeviceDefinitionOnChainService) {
	lis, err := net.Listen("tcp", ":"+s.GRPCPort)
	if err != nil {
		logger.Fatal().Msgf("Failed to listen on port %v: %v", s.GRPCPort, err)
	}

	deviceDefinitionService := NewGrpcService(m, &logger, dbs, onChainDeviceDefs)
	integrationService := NewGrpcIntegrationService(m, &logger)
	decodeService := NewGrpcVinDecoderService(m, &logger)

	logger.Info().Msgf("Starting gRPC server on port %s", s.GRPCPort)
	gp := common.GrpcConfig{Logger: &logger}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			metrics.ValidationMiddleware(),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(gp.GrpcConfig)),
		)),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	pkggrpc.RegisterDeviceDefinitionServiceServer(server, deviceDefinitionService)
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
