package prometheus

import (
	"fmt"

	promModel "github.com/prometheus/common/model"
)

const workloadsQueryTemplate = `
	sum(
		rate(
			istio_requests_total {
				reporter = "destination",
				source_app != "istio-ingressgateway",
				source_app != "telemetry",
				destination_app != "telemetry",
				source_app != "policy",
				destination_app != "policy",
				source_app != "mixer",
				destination_app != "mixer"
			}[%s]
		)
	) by (
		source_workload,
		destination_workload,
		source_app,
		destination_app
	)
`

// GetRequestsTotalByWorkloads returns request totals by workloads
func GetRequestsTotalByWorkloads(addr string) (promModel.Vector, error) {
	query := GetRequestsTotalByWorkloadsQuery()
	return executeQuery(addr, query)
}

// GetRequestsTotalByWorkloadsQuery returns request totals by workloads query
func GetRequestsTotalByWorkloadsQuery() string {
	return fmt.Sprintf(workloadsQueryTemplate, "60s")
}
