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
		id, workload = getWorkloadByMetric(metric, "source", workloads)

		// Add destination workload
		destinationWorkload := Workload{
			Name: string(metric["destination_workload"]),
			App:  string(metric["destination_app"]),
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
		id, workload = getWorkloadByMetric(
			metric,
			"destination",
			workloads,
		)

		// Add source workload
		sourceWorkload := Workload{
			Name: string(metric["source_workload"]),
			App:  string(metric["source_app"]),
		}
		workload.AddSource(sourceWorkload)

		workloads[id] = workload
	}

	return workloads, nil
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
		name = string(metric["source_workload"])
		app = string(metric["source_app"])
	} else {
		name = string(metric["destination_workload"])
		app = string(metric["destination_app"])
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
