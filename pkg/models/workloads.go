package models

import (
	"github.com/hekike/outlier-istio/pkg/prometheus"
	promModel "github.com/prometheus/common/model"
)

// GetWorkloads returns workload with it's destination workloads
func GetWorkloads(addr string) (map[string]Workload, error) {
	// Fetch data
	matrix, err := prometheus.GetRequestsTotalByWorkloads(addr)
	if err != nil {
		return nil, err
	}

	workloads := make(map[string]Workload)

	// Add sources with destinations
	for _, sample := range matrix {
		metric := sample.Metric

		// Get workload
		var id string
		var workload Workload
		id, workload = getSourceWorkloadByMetric(metric, workloads)

		// Add destination workload
		name, app := getDestinationFromMetric(metric)
		destinationWorkload := Workload{
			Name: name,
			App:  app,
		}
		workload.AddDestination(destinationWorkload)

		workloads[id] = workload
	}

	// Add destinations with sources
	for _, sample := range matrix {
		metric := sample.Metric
		// Get workload
		var id string
		var workload Workload
		id, workload = getDestinationWorkloadByMetric(metric, workloads)

		// Add source workload
		name, app := getSourceFromMetric(metric)
		sourceWorkload := Workload{
			Name: name,
			App:  app,
		}
		workload.AddSource(sourceWorkload)

		workloads[id] = workload
	}

	return workloads, nil
}

func getSourceWorkloadByMetric(metric promModel.Metric, workloads map[string]Workload) (
	id string,
	workload Workload,
) {
	return getWorkloadByMetric(metric, "source", workloads)
}

func getDestinationWorkloadByMetric(metric promModel.Metric, workloads map[string]Workload) (
	id string,
	workload Workload,
) {
	return getWorkloadByMetric(metric, "destination", workloads)
}

func getWorkloadByMetric(
	metric promModel.Metric,
	sourceType string,
	workloads map[string]Workload,
) (
	id string,
	workload Workload,
) {
	// Extract data
	var name, app string
	if sourceType == "source" {
		name, app = getSourceFromMetric(metric)
	} else {
		name, app = getDestinationFromMetric(metric)
	}

	// Find or create workload
	id = name + "-" + app

	if v, found := workloads[id]; found {
		workload = v
	} else {
		workload = Workload{
			Name:         name,
			App:          app,
			Sources:      make([]Workload, 0),
			Destinations: make([]Workload, 0),
		}
	}
	return id, workload
}
