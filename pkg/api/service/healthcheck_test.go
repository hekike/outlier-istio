package api

import (
	"context"
	"testing"

	"github.com/hekike/outlier-istio/pkg/api/protobuf"
	"github.com/stretchr/testify/assert"
)

func HealthcheckServiceTest(t *testing.T) {
	s := HealthServer{}

	req := &grpc_health_v1.HealthCheckRequest{}
	resp, err := s.Check(context.Background(), req)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, resp.Status, grpc_health_v1.HealthCheckResponse_SERVING)
}
