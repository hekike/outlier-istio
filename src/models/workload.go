package models

import (
	"context"
	"fmt"
	"time"

	promApi "github.com/prometheus/client_golang/api"
	promApiV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promModel "github.com/prometheus/common/model"
)

const workloadQuery = "sum(rate(istio_requests_total{reporter=\"destination\"}[%s])) by (source_workload, destination_workload, source_app, destination_app)"
const appTypeMixer = "mixer"

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

func (w *WorkloadItem) isMixer() bool {
	return w.App == appTypeMixer
}

func (w *Workload) addSource(wi WorkloadItem) []WorkloadItem {
	w.Sources = append(w.Sources, wi)
	return w.Sources
}

func (w *Workload) addDestination(wi WorkloadItem) []WorkloadItem {
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

		if workload.isMixer() {
			continue
		}

		// Add destination workload
		destinationWorkload := WorkloadItem{
			Name: string(metric["destination_workload"]),
			App:  string(metric["destination_app"]),
		}
		if !destinationWorkload.isMixer() {
			workload.addDestination(destinationWorkload)
		}

		workloads[id] = workload
	}

	// Add destinations with sources
	for _, sample := range matrix {
		metric := sample.Metric
		// Get workload
		var id string
		var workload Workload
		id, workload = getWorkloadByMetric(metric, "destination", workloads)

		if workload.isMixer() {
			continue
		}

		// Add source workload
		sourceWorkload := WorkloadItem{
			Name: string(metric["source_workload"]),
			App:  string(metric["source_app"]),
		}
		if !sourceWorkload.isMixer() {
			workload.addSource(sourceWorkload)
		}

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
