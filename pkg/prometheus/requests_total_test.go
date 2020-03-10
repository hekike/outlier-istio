package prometheus

import (
	"testing"

	"github.com/hekike/outlier-istio/test/fixtures"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

func TestGetRequestsTotalByWorkloads(t *testing.T) {
	mockServer := fixtures.PrometheusResponseStub(t, map[string]string{
		GetRequestsTotalByWorkloadsQuery(): "../../test/mock/prom_workload_request_totals.json",
	})
	defer mockServer.Close()

	result, err := GetRequestsTotalByWorkloads(mockServer.URL)
	if err != nil {
		t.Error(err)
	}

	expected := model.Vector{
		&model.Sample{
			Metric: model.Metric{
				"destination_app":      "reviews",
				"destination_workload": "reviews-v3",
				"request_protocol":     "http",
				"source_app":           "productpage",
				"source_workload":      "productpage-v1",
			},
			Timestamp: 1539917345608,
			Value:     0,
		},
		&model.Sample{
			Metric: model.Metric{
				"destination_app":      "productpage",
				"destination_workload": "productpage-v1",
				"request_protocol":     "http",
				"source_app":           "unknown",
				"source_workload":      "unknown",
			},
			Timestamp: 1539917345608,
			Value:     0,
		},
		&model.Sample{
			Metric: model.Metric{
				"destination_app":      "ratings",
				"destination_workload": "ratings-v1",
				"request_protocol":     "http",
				"source_app":           "reviews",
				"source_workload":      "reviews-v3",
			},
			Timestamp: 1539917345608,
			Value:     0,
		},
	}
	assert.Equal(t, expected, result)
}
