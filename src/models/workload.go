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

const destinationWorkloadRequestDurationPercentiles = `
	histogram_quantile(
		0.95,
		sum(
			rate(
				istio_request_duration_seconds_bucket {
					reporter = "destination",
					source_workload = "%s",
					destination_app != "mixer"
				}[%s]
			)
		) by (
			le,
			request_protocol,
			source_workload,
			source_app,
			destination_workload,
			destination_app
		)
	)
`

// WorkloadStatus struct.
type WorkloadStatus struct {
	Time   time.Time `json:"date"`
	Status string    `json:"status"`
}

// Workload struct.
type Workload struct {
	Name         string           `json:"name"`          // name of the workload
	App          string           `json:"app,omitempty"` // istio app
	Sources      []Workload       `json:"sources"`
	Destinations []Workload       `json:"destinations"`
	Statuses     []WorkloadStatus `json:"statuses,omitempty"`
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
	query := fmt.Sprintf(workloadQuery, "60s")
	matrix, err := fetchQuery(addr, query)
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
		id, workload = getWorkloadByMetric(metric, "destination", workloads)

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

// GetWorkloadStatusByName returns a single workload with it's status
func GetWorkloadStatusByName(addr string, name string) (*Workload, error) {
	// Fetch data
	query := fmt.Sprintf(
		destinationWorkloadRequestDurationPercentiles,
		"productpage-v1",
		"60s",
	)
	matrix, err := fetchQueryRange(addr, query)

	if err != nil {
		return nil, err
	}

	// Process data
	workload := Workload{}
	workload.Name = name
	workload.Statuses = make([]WorkloadStatus, 0)

	for _, sampleStream := range matrix {
		metric := sampleStream.Metric
		values := sampleStream.Values

		destinationWorkload := Workload{
			Name:     string(metric["destination_workload"]),
			App:      string(metric["destination_app"]),
			Statuses: make([]WorkloadStatus, 0),
		}

		for _, samplePair := range values {
			status := "unknown"

			if samplePair.Value > 0 {
				status = "ok"
			}

			t := samplePair.Timestamp.Time()
			workloadStatus := WorkloadStatus{
				Time:   t,
				Status: status,
			}

			destinationWorkload.Statuses = append(
				destinationWorkload.Statuses,
				workloadStatus,
			)

		}

		workload.AddDestination(destinationWorkload)
	}

	return &workload, nil
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

func fetchQueryRange(addr string, pq string) (promModel.Matrix, error) {
	client, err := promApi.NewClient(promApi.Config{Address: addr})
	if err != nil {
		return nil, err
	}
	api := promApiV1.NewAPI(client)

	// Query range
	queryTime := time.Now()
	queryRange := promApiV1.Range{
		Start: queryTime.Add(10 * -time.Minute),
		End:   queryTime,
		Step:  time.Minute,
	}
	val, err := api.QueryRange(context.Background(), pq, queryRange)
	if err != nil {
		return nil, err
	}
	matrix := val.(promModel.Matrix)

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
