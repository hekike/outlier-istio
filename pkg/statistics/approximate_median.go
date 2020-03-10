package statistics

import "sort"

// ApproximateMedian TODO: normal median for now
func ApproximateMedian(values Measurements) float64 {
	sort.Sort(values)

	length := len(values)
	middle := length / 2

	if length%2 == 1 {
		return values[middle]
	}

	return (values[middle-1] + values[middle]) / 2
}
