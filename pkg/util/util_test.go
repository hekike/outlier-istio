package util

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceInt64(t *testing.T) {
	values := SliceInt64{2, 1, 3, 4}
	sort.Sort(values)

	assert.Equal(t, SliceInt64{1, 2, 3, 4}, values)
}
