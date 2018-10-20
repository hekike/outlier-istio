package models

import (
	"context"
	"fmt"
	"time"

	promApi "github.com/prometheus/client_golang/api"
	promApiV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promModel "github.com/prometheus/common/model"
)

const workloadQuery = `
	sum(
		rate(
			istio_requests_total {
				reporter = "destination",
				source_app != "mixer",
				destination_app != "mixer"
			}[%s]
		)
	) by (
		source_workload,
		destination_workload,
		source_app,
		destination_app
	)
`

// WorkloadItem struct.
type WorkloadItem struct {
	Name string `json:"name"` // name of the workload
	App  string `json:"app"`  // istio app
}

// Workload struct.
type Workload struct {
	WorkloadItem
	Sources      []WorkloadItem `json:"sources"`
	Destinations []WorkloadItem `json:"destinations"`
}

// AddSource adds a source workload
func (w *Workload) AddSource(wi WorkloadItem) []WorkloadItem {
	w.Sources = append(w.Sources, wi)
	return w.Sources
}

// AddDestination adds a destination workload
func (w *Workload) AddDestination(wi WorkloadItem) []WorkloadItem {
	w.Destinations = append(w.Destinations, wi)
	return w.Destinations
}

// GetWorkloads returns workload with it's destination workloads
// TODO: refactor to use the same logic for adding source and destination
func GetWorkloads(addr string) (map[string]Workload, error) {
	// Fetch data
	matrix, err := fetchWorkloads(addr)
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
		destinationWorkload := WorkloadItem{
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
		id, workload = getWorkloadByMetric(metric, "destination", workloads)

		// Add source workload
		sourceWorkload := WorkloadItem{
			Name: string(metric["source_workload"]),
			App:  string(metric["source_app"]),
		}
		workload.AddSource(sourceWorkload)

		workloads[id] = workload
	}

	return workloads, nil
}

func fetchWorkloads(addr string) (promModel.Vector, error) {
	client, err := promApi.NewClient(promApi.Config{Address: addr})
	if err != nil {
		return nil, err
	}
	api := promApiV1.NewAPI(client)
	query := fmt.Sprintf(workloadQuery, "60s")

	val, err := api.Query(context.Background(), query, time.Now())
	if err != nil {
		return nil, err
	}
	matrix := val.(promModel.Vector)

	return matrix, nil
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
		workload = Workload{}
		workload.Name = name
		workload.App = app
		workload.Sources = make([]WorkloadItem, 0)
		workload.Destinations = make([]WorkloadItem, 0)
	}
	return id, workload
}
