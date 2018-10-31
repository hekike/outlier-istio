package models

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hekike/outlier-istio/src/utils"
	"github.com/montanaflynn/stats"
	promApi "github.com/prometheus/client_golang/api"
	promApiV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promModel "github.com/prometheus/common/model"
)

// workloadRequestDurationPercentiles returns a workload perentiles.
const workloadRequestDurationPercentiles = `
	histogram_quantile(
		0.95,
		sum(
			rate(
				istio_request_duration_seconds_bucket {
					reporter = "destination",
					%s_workload = "%s",
					destination_app != "mixer",
					destination_app != "telemetry",
					destination_app != "policy"
				}[%s]
			)
		) by (
			le,
			%s
		)
	)
`

// 0.1 millisecond accuracy (results are in second)
const decimals = 10000

// data resolution in Prometheus (Istio default is 5s)
const resolutionStep = 5 * time.Second

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

type sliceInt64 []int64

func (a sliceInt64) Len() int           { return len(a) }
func (a sliceInt64) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sliceInt64) Less(i, j int) bool { return a[i] < a[j] }

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
	hsv utils.SliceFloat64,
) []AggregatedStatusItem {
	statusItems := make([]AggregatedStatusItem, 0, len(as.Status))

	// Sort statuses by time
	keys := sliceInt64{}
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

		avg := utils.Avg(statusItem.Values)

		median, err := stats.Median(statusItem.Values)
		if err != nil {
			panic(err)
		}

		// Calculate approximate median and add current window's values
		// for the moving window
		var am float64

		if len(hsv) > 5 {
			am = utils.ApproximateMedian(hsv)
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

		if medianFormatted > amFormatted {
			statusItem.Status = "high"
		} else {
			statusItem.Status = "ok"
		}

		statusItems = append(statusItems, statusItem)
	}
	return statusItems
}

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
	workload := Workload{}
	workload.Name = name

	// Channels
	destination := make(chan workloadsResult)
	source := make(chan workloadsResult)
	workloadStatuses := make(chan workloadStatusesResult)

	// Fetch data by source
	go func() {
		result := workloadsResult{}

		matrixBySource, err := fetchQueryRange(
			addr,
			historicalStart,
			end,
			GetStatusQueryBySource(),
		)
		if err != nil {
			result.Error = err
			destination <- result
			return
		}

		// Iterate on destination workload dimension
		for _, sampleStream := range matrixBySource {
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
		destination <- result
		close(destination)
	}()

	// Fetch data by destination
	go func() {
		result := workloadsResult{}

		matrixByDestination, err := fetchQueryRange(
			addr,
			historicalStart,
			end,
			GetStatusQueryByDestination(),
		)
		if err != nil {
			result.Error = err
			source <- result
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
		source <- result
		close(source)
	}()

	// Workload status (aggregated destination)
	go func() {
		result := workloadStatusesResult{}

		matrix, err := fetchQueryRange(
			addr,
			historicalStart,
			end,
			GetStatusQuery(),
		)
		if err != nil {
			result.Error = err
			workloadStatuses <- result
			return
		}

		statuses := getWorkloadBySampleStream(
			matrix[0],
			start,
			statusStep,
		)

		result.StatusItems = statuses
		workloadStatuses <- result
	}()

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

// GetStatusQueryBySource returns a query
func GetStatusQueryBySource() string {
	return fmt.Sprintf(
		workloadRequestDurationPercentiles,
		"source",
		"productpage-v1",
		"60s",
		"request_protocol, source_workload, source_app, "+
			"destination_workload, destination_app",
	)
}

// GetStatusQueryByDestination returns a query
func GetStatusQueryByDestination() string {
	return fmt.Sprintf(
		workloadRequestDurationPercentiles,
		"destination",
		"productpage-v1",
		"60s",
		"request_protocol, source_workload, source_app, "+
			"destination_workload, destination_app",
	)
}

// GetStatusQuery returns a query
func GetStatusQuery() string {
	return fmt.Sprintf(
		workloadRequestDurationPercentiles,
		"source",
		"productpage-v1",
		"60s",
		"request_protocol",
	)
}

func getWorkloadBySampleStream(
	sampleStream *promModel.SampleStream,
	start time.Time,
	statusStep time.Duration,
) []AggregatedStatusItem {
	values := sampleStream.Values
	historicalSampleValues := utils.SliceFloat64{}

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
		Step:  resolutionStep,
	}
	val, err := api.QueryRange(context.Background(), pq, queryRange)
	if err != nil {
		return nil, err
	}
	matrix := val.(promModel.Matrix)

	return matrix, nil
}
