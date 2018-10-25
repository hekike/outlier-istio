package models

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/hekike/outlier-istio/src/utils"
	"github.com/montanaflynn/stats"
	promApi "github.com/prometheus/client_golang/api"
	promApiV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promModel "github.com/prometheus/common/model"
)

const destinationWorkloadRequestDurationPercentiles = `
	histogram_quantile(
		0.95,
		sum(
			rate(
				istio_request_duration_seconds_bucket {
					reporter = "destination",
					source_workload = "%s",
					destination_app != "mixer",
					destination_app != "telemetry",
					destination_app != "policy"
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

// AggregatedStatus calculates status.
type AggregatedStatus struct {
	Status map[int64]AggregatedStatusItem
}

// AggregatedStatusItem holds the status.
type AggregatedStatusItem struct {
	Time              time.Time `json:"date"`
	Status            string    `json:"status"`
	ApproximateMedian float64   `json:"approximateMedian"`
	Avg               float64   `json:"avg"`
	Median            float64   `json:"median"`
	Values            []float64 `json:"-"`
}

type unixTimeRange []int64

func (a unixTimeRange) Len() int           { return len(a) }
func (a unixTimeRange) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a unixTimeRange) Less(i, j int) bool { return a[i] < a[j] }

type sampleValues []float64

// AddStatus adds a new workload status.
func (as *AggregatedStatus) AddStatus(
	t time.Time,
	v float64,
) map[int64]AggregatedStatusItem {
	var statusItem AggregatedStatusItem
	roundedTime := t.Round(time.Minute)
	key := roundedTime.Unix()

	if v, found := as.Status[key]; found {
		statusItem = v
	} else {
		statusItem = AggregatedStatusItem{
			Time:   roundedTime,
			Status: "unknown",
		}
	}

	// TODO: is it valid to skip?
	if !math.IsNaN(v) {
		statusItem.Values = append(statusItem.Values, v)
	}

	as.Status[key] = statusItem
	return as.Status
}

// Aggregate turns the map to an aggregated array.
func (as *AggregatedStatus) Aggregate(hsv sampleValues) []AggregatedStatusItem {
	statusItems := make([]AggregatedStatusItem, 0, len(as.Status))

	keys := unixTimeRange{}
	for k := range as.Status {
		keys = append(keys, k)
	}
	sort.Sort(keys)

	// To perform the opertion you want
	for _, k := range keys {
		statusItem := as.Status[k]

		if len(statusItem.Values) < 3 {
			statusItems = append(statusItems, statusItem)
			continue
		}

		am := utils.ApproximateMedian(hsv)
		avg := utils.Avg(statusItem.Values)
		median, err := stats.Median(statusItem.Values)

		if err != nil {
			panic(err)
		}

		statusItem.ApproximateMedian = math.Round(am*1000) / 1000
		statusItem.Avg = math.Round(avg*1000) / 1000
		statusItem.Median = math.Round(median*1000) / 1000

		if statusItem.Median > statusItem.ApproximateMedian {
			statusItem.Status = "high"
		} else {
			statusItem.Status = "ok"
		}

		hsv = append(hsv, median)
		statusItems = append(statusItems, statusItem)
	}
	return statusItems
}

// GetWorkloadStatusByName returns a single workload with it's status.
func GetWorkloadStatusByName(
	addr string,
	name string,
	start time.Time,
	end time.Time,
) (*Workload, error) {
	// Fetch data
	query := fmt.Sprintf(
		destinationWorkloadRequestDurationPercentiles,
		"productpage-v1",
		"60s",
	)

	historicalStart := start.Add(-10 * time.Minute)

	matrix, err := fetchQueryRange(addr, historicalStart, end, query)

	if err != nil {
		return nil, err
	}

	// Process data
	workload := Workload{}
	workload.Name = name

	// Iterate on destination workload dimension
	for _, sampleStream := range matrix {
		metric := sampleStream.Metric
		values := sampleStream.Values
		historicalSampleValues := sampleValues{}

		aggregatedStatus := AggregatedStatus{
			Status: make(map[int64]AggregatedStatusItem),
		}

		// Sort sample pairs
		sort.Slice(values, func(i, j int) bool {
			return values[i].Timestamp.Time().Unix() < values[j].Timestamp.Time().Unix()
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

		destinationWorkload := Workload{
			Name:     string(metric["destination_workload"]),
			App:      string(metric["destination_app"]),
			Statuses: aggregatedStatus.Aggregate(historicalSampleValues),
		}

		workload.AddDestination(destinationWorkload)
	}

	return &workload, nil
}

func fetchQueryRange(
	addr string,
	start time.Time,
	end time.Time,
	pq string,
) (promModel.Matrix, error) {
	client, err := promApi.NewClient(promApi.Config{Address: addr})
	if err != nil {
		return nil, err
	}
	api := promApiV1.NewAPI(client)

	// Query range
	queryRange := promApiV1.Range{
		Start: start,
		End:   end,
		Step:  5 * time.Second,
	}
	val, err := api.QueryRange(context.Background(), pq, queryRange)
	if err != nil {
		return nil, err
	}
	matrix := val.(promModel.Matrix)

	return matrix, nil
}
