package models

import (
	"fmt"
	"math"
	"sort"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hekike/outlier-istio/pkg/prometheus"
	"github.com/hekike/outlier-istio/pkg/statistics"
	promModel "github.com/prometheus/common/model"
)

type workloadsResult struct {
	Workloads []Workload
	Error     error
}

type workloadStatusesResult struct {
	StatusItems []AggregatedStatusItem
	Error       error
}

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
		Statuses:     make([]AggregatedStatusItem, 0),
		Sources:      make([]Workload, 0),
		Destinations: make([]Workload, 0),
	}
	workload.Name = name

	// Fetch data by source
	destination := getStatusBySource(
		addr,
		historicalStart,
		end,
		statusStep,
		workload.Name,
	)

	// Fetch data by destination
	source := getStatusByDestination(
		addr,
		historicalStart,
		end,
		statusStep,
		workload.Name,
	)

	// Workload status (aggregated destination)
	workloadStatuses := getStatus(
		addr,
		historicalStart,
		end,
		statusStep,
		workload.Name)

	var err error

	// Source result
	sourceResult := <-source

	if sourceResult.Error != nil {
		err = multierror.Append(err, sourceResult.Error)
	} else {
		for _, w := range sourceResult.Workloads {
			workload.AddSource(w)
		}
	}

	// Destination result
	destinationResult := <-destination
	if destinationResult.Error != nil {
		err = multierror.Append(err, destinationResult.Error)
	} else {
		for _, w := range destinationResult.Workloads {
			workload.AddDestination(w)
		}
	}

	// Status result
	statusResult := <-workloadStatuses
	if statusResult.Error != nil {
		err = multierror.Append(err, statusResult.Error)
	} else {
		workload.Statuses = statusResult.StatusItems
	}

	return &workload, err
}

// Get status by source
func getStatusBySource(
	addr string,
	start time.Time,
	end time.Time,
	statusStep time.Duration,
	workload string,
) chan workloadsResult {
	results := make(chan workloadsResult)
	result := workloadsResult{}

	go func() {
		matrix, err := prometheus.GetStatusBySource(
			addr,
			start,
			end,
			workload,
		)
		if err != nil {
			result.Error = err
			results <- result
			return
		}

		// Iterate on destination workload dimension
		for _, sampleStream := range matrix {
			metric := sampleStream.Metric
			statuses := getWorkloadBySampleStream(
				sampleStream,
				start,
				statusStep,
			)

			destinationWorkload := Workload{
				Name:     string(metric["destination_workload"]),
				App:      string(metric["destination_app"]),
				Statuses: statuses,
			}

			result.Workloads = append(
				result.Workloads,
				destinationWorkload,
			)
		}
		results <- result
	}()

	return results
}

// Get status by destination
func getStatusByDestination(
	addr string,
	start time.Time,
	end time.Time,
	statusStep time.Duration,
	workload string,
) chan workloadsResult {
	results := make(chan workloadsResult)
	result := workloadsResult{}

	go func() {
		matrixByDestination, err := prometheus.GetStatusByDestination(
			addr,
			start,
			end,
			workload,
		)
		if err != nil {
			result.Error = err
			results <- result
			return
		}

		// Iterate on source workload dimension
		for _, sampleStream := range matrixByDestination {
			metric := sampleStream.Metric
			statuses := getWorkloadBySampleStream(
				sampleStream,
				start,
				statusStep,
			)

			sourceWorkload := Workload{
				Name:     string(metric["source_workload"]),
				App:      string(metric["source_app"]),
				Statuses: statuses,
			}

			result.Workloads = append(
				result.Workloads,
				sourceWorkload,
			)
		}
		results <- result
	}()

	return results
}

func getStatus(
	addr string,
	start time.Time,
	end time.Time,
	statusStep time.Duration,
	workload string,
) chan workloadStatusesResult {
	results := make(chan workloadStatusesResult)

	go func() {
		result := workloadStatusesResult{}

		matrix, err := prometheus.GetStatus(
			addr,
			start,
			end,
			workload,
		)
		if err != nil {
			result.Error = err
			results <- result
			return
		}

		fmt.Println(workload)

		if len(matrix) > 0 {
			statuses := getWorkloadBySampleStream(
				matrix[0],
				start,
				statusStep,
			)
			result.StatusItems = statuses
		} else {
			result.StatusItems = make([]AggregatedStatusItem, 0)
		}
		results <- result
	}()

	return results
}

func getWorkloadBySampleStream(
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
