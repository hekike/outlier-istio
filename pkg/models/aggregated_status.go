package models

import (
	"math"
	"sort"
	"time"

	"github.com/hekike/outlier-istio/pkg/statistics"
	"github.com/hekike/outlier-istio/pkg/util"
	"github.com/montanaflynn/stats"
)

// 0.1 millisecond accuracy (results are in second)
const decimals = 10000

// 0.5 milliseconds
const highTolerance = 0.5

// AggregatedStatus calculates status.
type AggregatedStatus struct {
	Step   time.Duration
	Status map[int64]AggregatedStatusItem
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

// AddStatus adds a new workload status.
func (as *AggregatedStatus) AddStatus(
	t time.Time,
	v float64,
) map[int64]AggregatedStatusItem {
	var statusItem AggregatedStatusItem
	roundedTime := t.Round(as.Step)
	key := roundedTime.Unix()

	if _statusItem, found := as.Status[key]; found {
		statusItem = _statusItem
	} else {
		statusItem = AggregatedStatusItem{
			Time:   roundedTime,
			Status: "unknown",
			Values: make([]float64, 0),
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
func (as *AggregatedStatus) Aggregate(
	hsv statistics.Measurements,
) []AggregatedStatusItem {
	statusItems := make([]AggregatedStatusItem, 0, len(as.Status))

	// Sort statuses by time
	keys := util.SliceInt64{}
	for k := range as.Status {
		keys = append(keys, k)
	}
	sort.Sort(keys)

	// Process statuses
	for _, k := range keys {
		statusItem := as.Status[k]

		if len(statusItem.Values) == 0 {
			statusItems = append(statusItems, statusItem)
			continue
		}

		avg := statistics.Avg(statusItem.Values)

		median, err := stats.Median(statusItem.Values)
		if err != nil {
			panic(err)
		}

		// Calculate approximate median and add current window's values
		// for the moving window
		var am float64

		if len(hsv) > 5 {
			am = statistics.ApproximateMedian(hsv)
		}
		for _, v := range statusItem.Values {
			hsv = append(hsv, v)
		}

		amFormatted := math.Round(am*decimals) / decimals
		avgFormatted := math.Round(avg*decimals) / decimals
		medianFormatted := math.Round(median*decimals) / decimals

		if !math.IsNaN(amFormatted) {
			statusItem.ApproximateMedian = &amFormatted
		}
		if !math.IsNaN(avgFormatted) {
			statusItem.Avg = &avgFormatted
		}
		if !math.IsNaN(medianFormatted) {
			statusItem.Median = &medianFormatted
		}

		if (medianFormatted - highTolerance) <= amFormatted {
			statusItem.Status = "ok"
		} else {
			statusItem.Status = "high"
		}

		statusItems = append(statusItems, statusItem)
	}
	return statusItems
}
