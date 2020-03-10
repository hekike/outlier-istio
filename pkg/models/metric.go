package models

import promModel "github.com/prometheus/common/model"

func getSourceFromMetric(metric promModel.Metric) (name string, app string) {
	name = string(metric["source_workload"])
	app = string(metric["source_app"])
	return name, app
}

func getDestinationFromMetric(metric promModel.Metric) (name string, app string) {
	name = string(metric["destination_workload"])
	app = string(metric["destination_app"])
	return name, app
}
