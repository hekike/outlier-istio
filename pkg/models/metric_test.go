package models

import (
	"testing"

	promModel "github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

func TestGetSourceFromMetric(t *testing.T) {
	metric := promModel.Metric{
		"source_workload": "workload-name",
		"source_app":      "workload-app",
	}
	name, app := getSourceFromMetric(metric)
	assert.Equal(t, name, "workload-name")
	assert.Equal(t, app, "workload-app")
}

func TestGetDestinationFromMetric(t *testing.T) {
	metric := promModel.Metric{
		"destination_workload": "workload-name",
		"destination_app":      "workload-app",
	}
	name, app := getDestinationFromMetric(metric)
	assert.Equal(t, name, "workload-name")
	assert.Equal(t, app, "workload-app")
}
