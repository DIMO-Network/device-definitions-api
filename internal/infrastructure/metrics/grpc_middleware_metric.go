package metrics

import (
	"context"

	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func ValidationMiddleware() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		resp, err := handler(ctx, req)

		if err != nil {
			if s, ok := status.FromError(err); ok {
				GRPCResponseTime.With(prometheus.Labels{"method": info.FullMethod, "status": s.Code().String()}).Observe(time.Since(startTime).Seconds())
				GRPCRequestCount.With(prometheus.Labels{"method": info.FullMethod, "status": s.Code().String()}).Inc()
			} else {
				GRPCResponseTime.With(prometheus.Labels{"method": info.FullMethod, "status": "unknown"}).Observe(time.Since(startTime).Seconds())
				GRPCRequestCount.With(prometheus.Labels{"method": info.FullMethod, "status": "unknown"}).Inc()
			}
		} else {
			GRPCResponseTime.With(prometheus.Labels{"method": info.FullMethod, "status": "OK"}).Observe(time.Since(startTime).Seconds())
			GRPCRequestCount.With(prometheus.Labels{"method": info.FullMethod, "status": "OK"}).Inc()
		}

		return resp, err
	}
}
