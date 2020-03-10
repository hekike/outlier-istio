package statistics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApproximateMedian(t *testing.T) {
	input := Measurements{1.0, 2.0, 3.0}
	output := ApproximateMedian(input)
	assert.Equal(t, 2.0, output)

	input = Measurements{1.0, 2.0}
	output = ApproximateMedian(input)
	assert.Equal(t, 1.5, output)
}
