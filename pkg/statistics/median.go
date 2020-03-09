package statistics

import (
	"math"
)

// ApproximateMedian returns with an approximate median
// http://www.dmi.unict.it/~battiato/download/MedianLNCS.pdf
func ApproximateMedian(A Measurements) float64 {
	size := len(A)
	step := 1
	r := int(math.Round(root(float64(size), 3)))

	for j := 0; j <= r; j++ {
		i := (step - 1) / 2
		for i < size {
			A = tripletAdjut(A, i, step)
			i = i + (3 * step)
		}
		step = 3 * step
	}
	return A[(size-1)/2]
}

func tripletAdjut(A Measurements, i int, step int) Measurements {
	size := len(A)
	j := i + step
	k := i + 2

	if i >= size-1 || j >= size-1 {
		return A
	}

	if A[i] < A[j] {
		if A[k] < A[i] {
			A.Swap(i, j)
			return A
		} else if A[k] < A[j] {
			A.Swap(j, k)
			return A
		}
	} else {
		if A[i] < A[k] {
			A.Swap(i, j)
			return A
		} else if A[k] > A[j] {
			A.Swap(j, k)
			return A
		}
	}

	return A
}

func root(a float64, n int) float64 {
	n1 := n - 1
	n1f, rn := float64(n1), 1/float64(n)
	x, x0 := 1., 0.
	for {
		potx, t2 := 1/x, a
		for b := n1; b > 0; b >>= 1 {
			if b&1 == 1 {
				t2 *= potx
			}
			potx *= potx
		}
		x0, x = x, rn*(n1f*x+t2)
		if math.Abs(x-x0)*1e15 < x {
			break
		}
	}
	return x
}