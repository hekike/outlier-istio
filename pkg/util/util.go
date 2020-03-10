package util

// SliceInt64 is a Int64 utility
type SliceInt64 []int64

func (a SliceInt64) Len() int           { return len(a) }
func (a SliceInt64) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SliceInt64) Less(i, j int) bool { return a[i] < a[j] }
