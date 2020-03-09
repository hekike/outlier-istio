package models

import (
	"math"
	"sort"
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hekike/outlier-istio/pkg/prometheus"
	"github.com/hekike/outlier-istio/pkg/statistics"
	promModel "github.com/prometheus/common/model"
)

// GetWorkloadStatusByName returns a single workload with it's status.
func GetWorkloadStatusByName(
	addr string,
	name string,
	start time.Time,
	end time.Time,
	historical time.Duration,
	statusStep time.Duration,
) (*Workload, error) {
	historicalStart := start.Add(-historical)
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
		statuses := getStatusesBySampleStream(
			sampleStream,
			start,
			statusStep,
		)

		workload := Workload{
			Name:     string(metric["destination_workload"]),
			App:      string(metric["destination_app"]),
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
		statuses := getStatusesBySampleStream(
			sampleStream,
			start,
			statusStep,
		)

		workload := Workload{
			Name:     string(metric["source_workload"]),
			App:      string(metric["source_app"]),
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
		statuses := getStatusesBySampleStream(
			matrix[0],
			start,
			statusStep,
		)
		return statuses, nil
	}
	return make([]AggregatedStatusItem, 0), nil
}

// Calculates statuses based on samples
func getStatusesBySampleStream(
	sampleStream *promModel.SampleStream,
	start time.Time,
	statusStep time.Duration,
) []AggregatedStatusItem {
	values := sampleStream.Values
	historicalSampleValues := statistics.Measurements{}

	aggregatedStatus := AggregatedStatus{
		Step:   statusStep,
		Status: make(map[int64]AggregatedStatusItem),
	}

	// Sort sample pairs
	sort.Slice(values, func(i, j int) bool {
		return values[i].Timestamp.Time().Unix() <
			values[j].Timestamp.Time().Unix()
	})

	// Iterate on time dimension
	for _, samplePair := range values {
		t := samplePair.Timestamp.Time()
		v := float64(samplePair.Value)

		// Value in the current range
		if t.Unix() > start.Unix() {
			aggregatedStatus.AddStatus(t, v)
		} else {
			// Historical value
			// TODO: is it valid to skip?
			if !math.IsNaN(v) {
				historicalSampleValues = append(
					historicalSampleValues,
					v,
				)
			}
		}
	}
	statuses := aggregatedStatus.Aggregate(historicalSampleValues)
	return statuses
}
