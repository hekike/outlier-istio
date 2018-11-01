package models

import (
	"context"
	"fmt"
	"time"

	promApi "github.com/prometheus/client_golang/api"
	promApiV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promModel "github.com/prometheus/common/model"
)

// WorkloadQuery returns workloads
const workloadQuery = `
	sum(
		rate(
			istio_requests_total {
				reporter = "destination",
				source_app != "telemetry",
				destination_app != "telemetry",
				source_app != "policy",
				destination_app != "policy",
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

// WorkloadStatus struct.
type WorkloadStatus struct {
	Time   time.Time `json:"date"`
	Status string    `json:"status"`
}

// Workload struct.
type Workload struct {
	Name         string                 `json:"name"`          // name of the workload
	App          string                 `json:"app,omitempty"` // istio app
	Sources      []Workload             `json:"sources"`
	Destinations []Workload             `json:"destinations"`
	Statuses     []AggregatedStatusItem `json:"statuses,omitempty"`
}

// AddSource adds a source workload
func (w *Workload) AddSource(wi Workload) []Workload {
	w.Sources = append(w.Sources, wi)
	return w.Sources
}

// AddDestination adds a destination workload
func (w *Workload) AddDestination(wi Workload) []Workload {
	w.Destinations = append(w.Destinations, wi)
	return w.Destinations
}

// GetWorkloads returns workload with it's destination workloads
func GetWorkloads(addr string) (map[string]Workload, error) {
	// Fetch data
	matrix, err := fetchQuery(addr, GetWorkloadQuery())
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

// GetWorkloadQuery returns a workload query
func GetWorkloadQuery() string {
	return fmt.Sprintf(workloadQuery, "60s")
}

func fetchQuery(addr string, pq string) (promModel.Vector, error) {
	client, err := promApi.NewClient(promApi.Config{Address: addr})
	if err != nil {
		return nil, err
	}
	api := promApiV1.NewAPI(client)

	val, err := api.Query(context.Background(), pq, time.Now())
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
		workload = Workload{
			Name:         name,
			App:          app,
			Sources:      make([]Workload, 0),
			Destinations: make([]Workload, 0),
		}
	}
	return id, workload
}
