package models

import (
	"testing"
	"time"

	"github.com/hekike/outlier-istio/pkg/statistics"
	"github.com/stretchr/testify/assert"
)

func TestAddSample(t *testing.T) {
	status := AggregatedStatus{
		StatusTimeline: map[unixTime]AggregatedStatusItem{},
	}
	sampleTime1, _ := time.Parse(time.RFC3339, "1970-01-01T00:01:01.000001+00:00")
	var value1 float64 = 10.0

	sampleTime2, _ := time.Parse(time.RFC3339, "1970-01-01T00:01:00.000000+00:00")
	var value2 float64 = 12.0
	sampleTime3, _ := time.Parse(time.RFC3339, "1970-01-01T00:01:00.000001+00:00")
	var value3 float64 = 13.0

	status.AddSample(sampleTime1, value1)
	status.AddSample(sampleTime2, value2)
	status.AddSample(sampleTime3, value3)

	assert.Equal(t, AggregatedStatus{
		StatusTimeline: map[unixTime]AggregatedStatusItem{
			// should sort samples
			60: AggregatedStatusItem{
				Time:   sampleTime2,
				Values: []float64{value2, value3},
			},
			61: AggregatedStatusItem{
				Time:   sampleTime1,
				Values: []float64{value1},
			},
		},
	}, status)
}

func TestAggregate(t *testing.T) {
	sampleTime1, _ := time.Parse(time.RFC3339, "1970-01-01T00:01:00.000001+00:00")
	sampleTime2, _ := time.Parse(time.RFC3339, "1970-01-01T00:01:02.000000+00:00")

	status := AggregatedStatus{
		StatusTimeline: map[unixTime]AggregatedStatusItem{
			60: AggregatedStatusItem{
				Time:   sampleTime1,
				Values: []float64{10, 11, 12, 11, 13, 19},
			},
			61: AggregatedStatusItem{
				Time:   sampleTime2,
				Values: []float64{12, 14, 15, 16, 17},
			},
		},
	}

	var historicalSampleValues statistics.Measurements = statistics.Measurements{10, 11, 12, 13, 12, 11}
	statuses := status.Aggregate(historicalSampleValues)

	am1 := 12.0
	avg1 := 12.6667
	median1 := 11.5
	am2 := 11.0
	avg2 := 14.8
	median2 := 15.0
	assert.Equal(t, []AggregatedStatusItem{
		AggregatedStatusItem{
			Time:              sampleTime1,
			Status:            "ok",
			Values:            statistics.Measurements{10, 11, 12, 11, 13, 19},
			ApproximateMedian: &am1,
			Avg:               &avg1,
			Median:            &median1,
		},
		AggregatedStatusItem{
			Time:              sampleTime2,
			Status:            "high",
			Values:            statistics.Measurements{12, 14, 15, 16, 17},
			ApproximateMedian: &am2,
			Avg:               &avg2,
			Median:            &median2,
		},
	}, statuses)
}
