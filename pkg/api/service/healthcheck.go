package api

import (
	"context"

	"github.com/hekike/outlier-istio/pkg/api/protobuf"
)

// HealthServer is used to implement grpc_health_v1.HealthServer.
type HealthServer struct{}

// Check implements grpc_health_v1.Check
func (s *HealthServer) Check(
	ctx context.Context,
	in *grpc_health_v1.HealthCheckRequest,
) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}
