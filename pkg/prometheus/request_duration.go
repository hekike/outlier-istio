package prometheus

import (
	"fmt"
	"time"

	"github.com/prometheus/common/model"
)

const workloadRequestDurationPercentilesTemplate = `
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

// data resolution in Prometheus (Istio default is 5s)
const resolutionStep = 5 * time.Second

// GetDownstreamRequestDurations returns request durations for workloads called
// from the given workload.
func GetDownstreamRequestDurations(
	addr string,
	start time.Time,
	end time.Time,
	workload string,
) (model.Matrix, error) {
	query := GetDownstreamRequestDurationsQuery(workload)
	return executeQueryRange(addr, start, end, query)
}

// GetDownstreamRequestDurationsQuery returns a Prometheus query
func GetDownstreamRequestDurationsQuery(workload string) string {
	return fmt.Sprintf(
		workloadRequestDurationPercentilesTemplate,
		"source",
		workload,
		"60s",
		"request_protocol, source_workload, source_app, destination_workload, destination_app",
	)
}

// GetUpstreamRequestDurations returns request durations for requests made to
// given workload from sources.
func GetUpstreamRequestDurations(
	addr string,
	start time.Time,
	end time.Time,
	workload string,
) (model.Matrix, error) {
	query := GetUpstreamRequestDurationsQuery(workload)
	return executeQueryRange(addr, start, end, query)
}

// GetUpstreamRequestDurationsQuery returns a Prometheus query
func GetUpstreamRequestDurationsQuery(workload string) string {
	return fmt.Sprintf(
		workloadRequestDurationPercentilesTemplate,
		"destination",
		workload,
		"60s",
		"request_protocol, source_workload, source_app, destination_workload, destination_app",
	)
}

// GetStatuses returns statuses for given workload
func GetStatuses(
	addr string,
	start time.Time,
	end time.Time,
	workload string,
) (model.Matrix, error) {
	query := GetStatusesQuery(workload)
	return executeQueryRange(addr, start, end, query)
}

// GetStatusesQuery returns statuses query for given workload
func GetStatusesQuery(workload string) string {
	return fmt.Sprintf(
		workloadRequestDurationPercentilesTemplate,
		"destination",
		workload,
		"60s",
		"request_protocol",
	)
}
