package statistics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAvg(t *testing.T) {
	input := Measurements{10.5, 10.5}
	output := Avg(input)
	assert.Equal(t, 10.5, output)

	input = Measurements{10, 20}
	output = Avg(input)
	assert.Equal(t, 15.0, output)
}
