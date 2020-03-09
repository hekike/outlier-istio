package statistics

// Measurements contains multiple data points
type Measurements []float64

func (a Measurements) Len() int           { return len(a) }
func (a Measurements) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Measurements) Less(i, j int) bool { return a[i] < a[j] }
