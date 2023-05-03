package api

import (
	"context"
	"net"
	"time"

	"github.com/DIMO-Network/device-definitions-api/internal/api/common"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"
	pkggrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func recoveryMiddleware() grpc.UnaryServerInterceptor {
	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(common.GrpcConfig),
	}
	return grpc_recovery.UnaryServerInterceptor(recoveryOpts...)
}

func validationMiddleware() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		resp, err := handler(ctx, req)

		if err != nil {
			if s, ok := status.FromError(err); ok {
				metrics.GRPCResponseTime.With(prometheus.Labels{"method": info.FullMethod, "status": s.Code().String()}).Observe(time.Since(startTime).Seconds())
				metrics.GRPCRequestCount.With(prometheus.Labels{"method": info.FullMethod, "status": s.Code().String()}).Inc()
			} else {
				metrics.GRPCResponseTime.With(prometheus.Labels{"method": info.FullMethod, "status": "unknown"}).Observe(time.Since(startTime).Seconds())
				metrics.GRPCRequestCount.With(prometheus.Labels{"method": info.FullMethod, "status": "unknown"}).Inc()
			}
		} else {
			metrics.GRPCResponseTime.With(prometheus.Labels{"method": info.FullMethod, "status": "OK"}).Observe(time.Since(startTime).Seconds())
			metrics.GRPCRequestCount.With(prometheus.Labels{"method": info.FullMethod, "status": "OK"}).Inc()
		}

		return resp, err
	}
}

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
			validationMiddleware(),
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
