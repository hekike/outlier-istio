package models

import (
	"math"
	"sort"
	"time"

	"github.com/hekike/outlier-istio/pkg/statistics"
	"github.com/hekike/outlier-istio/pkg/util"
	"github.com/montanaflynn/stats"
	promModel "github.com/prometheus/common/model"
)

// 0.1 millisecond accuracy (results are in second)
const decimals = 10000

// 0.5 milliseconds
const highTolerance = 0.5

type unixTime = int64

// AggregatedStatus calculates status.
type AggregatedStatus struct {
	Step           time.Duration
	StatusTimeline map[unixTime]AggregatedStatusItem
}

// AggregatedStatusItem holds the status.
type AggregatedStatusItem struct {
	Time   time.Time `json:"date"`
	Status string    `json:"status"`
	Values []float64 `json:"-"`
	// pointer as they can be JSON null
	ApproximateMedian *float64 `json:"approximateMedian"`
	Avg               *float64 `json:"avg"`
	Median            *float64 `json:"median"`
}

// AddSample adds a new workload status.
func (as *AggregatedStatus) AddSample(
	t time.Time,
	v float64,
) {
	var statusItem AggregatedStatusItem

	// Map time to closest step
	roundedTime := t.Round(as.Step)
	timeKey := roundedTime.Unix()

	// Find or create status item
	if _statusItem, found := as.StatusTimeline[timeKey]; found {
		statusItem = _statusItem
	} else {
		statusItem = AggregatedStatusItem{
			Time:   roundedTime,
			Values: make([]float64, 0),
		}
	}

	// TODO: is it valid to skip?
	if !math.IsNaN(v) {
		statusItem.Values = append(statusItem.Values, v)
	}

	// Store status item
	as.StatusTimeline[timeKey] = statusItem
}

// Aggregate turns the map to an aggregated array.
func (as *AggregatedStatus) Aggregate(
	historicalSampleValues statistics.Measurements,
) []AggregatedStatusItem {
	statusItems := make([]AggregatedStatusItem, 0, len(as.StatusTimeline))

	// Sort timeline steps
	timeKeys := util.SliceInt64{}
	for timeKey := range as.StatusTimeline {
		timeKeys = append(timeKeys, timeKey)
	}
	sort.Sort(timeKeys)

	// Process statuses in time order
	for _, timeKey := range timeKeys {
		statusItem := as.StatusTimeline[timeKey]

		// Skip if we don't have any values for time frame
		if len(statusItem.Values) == 0 {
			statusItems = append(statusItems, statusItem)
			continue
		}

		// Statistics
		avg := statistics.Avg(statusItem.Values)
		median, err := stats.Median(statusItem.Values)
		if err != nil {
			panic(err)
		}

		// Calculate approximate median and add current window's values
		// for the moving window
		var approximateMedian float64

		if len(historicalSampleValues) > 5 {
			approximateMedian = statistics.ApproximateMedian(historicalSampleValues)
		}
		for _, value := range statusItem.Values {
			historicalSampleValues = append(historicalSampleValues, value)
		}

		// Store statistical results
		amFormatted := roundToDecimals(approximateMedian)
		if !math.IsNaN(amFormatted) {
			statusItem.ApproximateMedian = &amFormatted
		}
		avgFormatted := roundToDecimals(avg)
		if !math.IsNaN(avgFormatted) {
			statusItem.Avg = &avgFormatted
		}
		medianFormatted := roundToDecimals(median)
		if !math.IsNaN(medianFormatted) {
			statusItem.Median = &medianFormatted
		}

		// Determinate status
		if (medianFormatted - highTolerance) <= amFormatted {
			statusItem.Status = "ok"
		} else {
			statusItem.Status = "high"
		}

		statusItems = append(statusItems, statusItem)
	}
	return statusItems
}

// Calculates statuses based on samples
func calculateStatusesBySamples(
	samples []promModel.SamplePair,
	start time.Time,
	statusStep time.Duration,
) []AggregatedStatusItem {
	historicalSampleValues := statistics.Measurements{}

	aggregatedStatus := AggregatedStatus{
		Step:           statusStep,
		StatusTimeline: make(map[int64]AggregatedStatusItem),
	}

	// Sort sample pairs by time
	sort.Slice(samples, func(i, j int) bool {
		return samples[i].Timestamp.Time().Unix() <
			samples[j].Timestamp.Time().Unix()
	})

	// Iterate on time dimension
	for _, samplePair := range samples {
		time := samplePair.Timestamp.Time()
		value := float64(samplePair.Value)

		// Value in the current range
		if time.Unix() > start.Unix() {
			aggregatedStatus.AddSample(time, value)
		} else {
			// Historical value
			// TODO: is it valid to skip?
			if !math.IsNaN(value) {
				historicalSampleValues = append(
					historicalSampleValues,
					value,
				)
			}
		}
	}

	// Calculate statuses
	statuses := aggregatedStatus.Aggregate(historicalSampleValues)
	return statuses
}

func roundToDecimals(value float64) float64 {
	return math.Round(value*decimals) / decimals
}
