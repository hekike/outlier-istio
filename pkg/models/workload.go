package models

import (
	"time"
)

// WorkloadStatus struct.
type WorkloadStatus struct {
	Time   time.Time `json:"date"`
	Status string    `json:"status"`
}

// Workload struct.
type Workload struct {
	Name         string                 `json:"name"`          // name of the workload
	App          string                 `json:"app,omitempty"` // istio app
	Sources      []Workload             `json:"sources"`
	Destinations []Workload             `json:"destinations"`
	Statuses     []AggregatedStatusItem `json:"statuses"`
}

// AddSource adds a source workload
func (w *Workload) AddSource(wi Workload) []Workload {
	w.Sources = append(w.Sources, wi)
	return w.Sources
}

// AddDestination adds a destination workload
func (w *Workload) AddDestination(wi Workload) []Workload {
	w.Destinations = append(w.Destinations, wi)
	return w.Destinations
}
