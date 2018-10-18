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

func (aw *Workload) addSource(ds WorkloadItem) []WorkloadItem {
	aw.Sources = append(aw.Sources, ds)
	return aw.Sources
}

func (aw *Workload) addDestination(ds WorkloadItem) []WorkloadItem {
	aw.Destinations = append(aw.Destinations, ds)
	return aw.Destinations
}

// GetWorkloads returns workload with it's destination workloads
// TODO: refactor to use the same logic for adding source and destination
func GetWorkloads(addr string) (map[string]Workload, error) {
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

	workloads := make(map[string]Workload)
	for _, sample := range matrix {
		metrics := sample.Metric
		name := string(metrics["source_workload"])
		app := string(metrics["source_app"])

		if app == "mixer" {
			continue
		}

		var id string
		var workload Workload
		id, workload = getWorkload(name, app, workloads)

		// Add destination workload
		destinationWorkload := WorkloadItem{
			Name: string(metrics["destination_workload"]),
			App:  string(metrics["destination_app"]),
		}
		if destinationWorkload.App != "mixer" {
			workload.addDestination(destinationWorkload)
		}

		workloads[id] = workload
	}

	// Add sources
	for _, sample := range matrix {
		metrics := sample.Metric
		name := string(metrics["destination_workload"])
		app := string(metrics["destination_app"])

		if app == "mixer" {
			continue
		}

		var id string
		var workload Workload
		id, workload = getWorkload(name, app, workloads)

		// Add source workload
		sourceWorkload := WorkloadItem{
			Name: string(metrics["source_workload"]),
			App:  string(metrics["source_app"]),
		}
		if sourceWorkload.App != "mixer" {
			workload.addSource(sourceWorkload)
		}

		workloads[id] = workload
	}

	return workloads, nil
}

func getWorkload(
	name string,
	app string,
	workloads map[string]Workload,
) (
	id string,
	workload Workload,
) {
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
