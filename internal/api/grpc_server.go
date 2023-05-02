package api

import (
	"context"
	"net"
	"time"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/DIMO-Network/device-definitions-api/internal/api/common"
	"github.com/DIMO-Network/device-definitions-api/internal/config"
	"github.com/DIMO-Network/device-definitions-api/internal/infrastructure/metrics"
	pkggrpc "github.com/DIMO-Network/device-definitions-api/pkg/grpc"
	"github.com/TheFellow/go-mediator/mediator"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func metricsMiddleware() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		resp, err := handler(ctx, req)
		metrics.GRPCResponseTime.With(prometheus.Labels{"method": info.FullMethod, "status": status.Code(err).String()}).Observe(time.Since(startTime).Seconds())
		metrics.GRPCRequestCount.With(prometheus.Labels{"method": info.FullMethod, "status": status.Code(err).String()}).Inc()
		//metrics.ResponseTime.WithLabelValues(info.FullMethod).Observe(time.Since(startTime).Seconds())
		//metrics.RequestCount.WithLabelValues(info.FullMethod).Inc()
		return resp, err
	}
}

func StartGrpcServer(logger zerolog.Logger, s *config.Settings, m mediator.Mediator) {
	lis, err := net.Listen("tcp", ":"+s.GRPCPort)
	if err != nil {
		logger.Fatal().Msgf("Failed to listen on port %v: %v", s.GRPCPort, err)
	}

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(common.GrpcConfig),
	}

	deviceDefinitionService := NewGrpcService(m, &logger)
	recallsService := NewGrpcRecallsService(m, &logger)
	reviewsService := NewGrpcReviewsService(m, &logger)
	integrationService := NewGrpcIntegrationService(m, &logger)
	decodeService := NewGrpcVinDecoderService(m, &logger)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(opts...),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			metricsMiddleware(),
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
