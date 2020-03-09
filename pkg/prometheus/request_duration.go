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

// GetStatusBySource returns statuses by source
func GetStatusBySource(
	addr string,
	start time.Time,
	end time.Time,
	workload string,
) (model.Matrix, error) {
	query := GetStatusBySourceQuery(workload)
	return executeQueryRange(addr, start, end, query)
}

// GetStatusBySourceQuery returns statuses by source query
func GetStatusBySourceQuery(workload string) string {
	return fmt.Sprintf(
		workloadRequestDurationPercentilesTemplate,
		"source",
		workload,
		"60s",
		"request_protocol, source_workload, source_app, destination_workload, destination_app",
	)
}

// GetStatusByDestination returns statuses by destination
func GetStatusByDestination(
	addr string,
	start time.Time,
	end time.Time,
	workload string,
) (model.Matrix, error) {
	query := GetStatusByDestinationQuery(workload)
	return executeQueryRange(addr, start, end, query)
}

// GetStatusByDestinationQuery returns statuses by destination query
func GetStatusByDestinationQuery(workload string) string {
	return fmt.Sprintf(
		workloadRequestDurationPercentilesTemplate,
		"destination",
		workload,
		"60s",
		"request_protocol, source_workload, source_app, destination_workload, destination_app",
	)
}

// GetStatus returns a query
func GetStatus(
	addr string,
	start time.Time,
	end time.Time,
	workload string,
) (model.Matrix, error) {
	query := GetStatusQuery(workload)
	return executeQueryRange(addr, start, end, query)
}

// GetStatusQuery returns a status query
func GetStatusQuery(workload string) string {
	return fmt.Sprintf(
		workloadRequestDurationPercentilesTemplate,
		"destination",
		workload,
		"60s",
		"request_protocol",
	)
}
