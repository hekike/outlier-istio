package models

import (
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hekike/outlier-istio/pkg/prometheus"
)

// WorkloadStatus struct.
type WorkloadStatus struct {
	Time   time.Time `json:"date"`
	Status string    `json:"status"`
}

// GetWorkloadStatusByName returns a single workload with it's status.
func GetWorkloadStatusByName(
	addr string,
	name string,
	start time.Time,
	end time.Time,
	/** we fetch historical values for baseline calculation */
	historicalStart time.Time,
	statusStep time.Duration,
) (*Workload, error) {
	workload := Workload{
		Name:         name,
		Statuses:     make([]AggregatedStatusItem, 0),
		Sources:      make([]Workload, 0),
		Destinations: make([]Workload, 0),
	}

	var wg sync.WaitGroup
	var combinedErr error

	// Add destinations
	wg.Add((1))
	go func() {
		defer wg.Done()
		workloads, err := getDownstreams(
			addr,
			historicalStart,
			end,
			statusStep,
			workload.Name,
		)
		if err != nil {
			combinedErr = multierror.Append(combinedErr, err)
		}
		for _, w := range workloads {
			workload.AddDestination(w)
		}
	}()

	// Add sources
	wg.Add((1))
	go func() {
		defer wg.Done()
		workloads, err := getUpstreams(
			addr,
			historicalStart,
			end,
			statusStep,
			workload.Name,
		)
		if err != nil {
			combinedErr = multierror.Append(combinedErr, err)
		}
		for _, w := range workloads {
			workload.AddSource(w)
		}
	}()

	// Add aggregated statuses
	wg.Add((1))
	go func() {
		defer wg.Done()
		statuses, err := getStatuses(
			addr,
			historicalStart,
			end,
			statusStep,
			workload.Name,
		)
		if err != nil {
			combinedErr = multierror.Append(combinedErr, err)
		}
		workload.Statuses = statuses
	}()

	wg.Wait()

	return &workload, combinedErr
}

// Get downstream workloads with statuses
func getDownstreams(
	addr string,
	start time.Time,
	end time.Time,
	statusStep time.Duration,
	workload string,
) ([]Workload, error) {
	workloads := []Workload{}

	matrix, err := prometheus.GetDownstreamRequestDurations(
		addr,
		start,
		end,
		workload,
	)
	if err != nil {
		return workloads, err
	}

	// Iterate on destination workload dimension
	for _, sampleStream := range matrix {
		metric := sampleStream.Metric
		statuses := calculateStatusesBySamples(
			sampleStream.Values,
			start,
			statusStep,
		)
		name, app := getDestinationFromMetric(metric)

		workload := Workload{
			Name:     name,
			App:      app,
			Statuses: statuses,
		}

		workloads = append(
			workloads,
			workload,
		)
	}

	return workloads, nil
}

// Get upstream workloads with statuses
func getUpstreams(
	addr string,
	start time.Time,
	end time.Time,
	statusStep time.Duration,
	workload string,
) ([]Workload, error) {
	workloads := []Workload{}

	matrixByDestination, err := prometheus.GetUpstreamRequestDurations(
		addr,
		start,
		end,
		workload,
	)
	if err != nil {
		return workloads, err
	}

	// Iterate on source workload dimension
	for _, sampleStream := range matrixByDestination {
		metric := sampleStream.Metric
		statuses := calculateStatusesBySamples(
			sampleStream.Values,
			start,
			statusStep,
		)

		name, app := getSourceFromMetric(metric)
		workload := Workload{
			Name:     name,
			App:      app,
			Statuses: statuses,
		}

		workloads = append(
			workloads,
			workload,
		)
	}

	return workloads, nil
}

// Returns statuses for given workload
func getStatuses(
	addr string,
	start time.Time,
	end time.Time,
	statusStep time.Duration,
	workload string,
) ([]AggregatedStatusItem, error) {
	matrix, err := prometheus.GetStatuses(
		addr,
		start,
		end,
		workload,
	)
	if err != nil {
		return make([]AggregatedStatusItem, 0), err
	}

	if len(matrix) > 0 {
		statuses := calculateStatusesBySamples(
			matrix[0].Values,
			start,
			statusStep,
		)
		return statuses, nil
	}
	return make([]AggregatedStatusItem, 0), nil
}
