package statistics

// Avg calculates the average
func Avg(xs Measurements) float64 {
	total := 0.0
	for _, v := range xs {
		total += v
	}
	return total / float64(len(xs))
}
