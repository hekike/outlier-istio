package prometheus

import (
	"context"
	"time"

	promApi "github.com/prometheus/client_golang/api"
	promApiV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promModel "github.com/prometheus/common/model"
)

func executeQuery(addr string, pq string) (promModel.Vector, error) {
	client, err := promApi.NewClient(promApi.Config{Address: addr})
	if err != nil {
		return nil, err
	}
	api := promApiV1.NewAPI(client)

	val, _, err := api.Query(context.Background(), pq, time.Now())
	if err != nil {
		return nil, err
	}
	matrix := val.(promModel.Vector)

	return matrix, nil
}

func executeQueryRange(
	addr string,
	start time.Time,
	end time.Time,
	pq string,
) (promModel.Matrix, error) {
	client, err := promApi.NewClient(promApi.Config{Address: addr})
	if err != nil {
		return nil, err
	}
	api := promApiV1.NewAPI(client)

	// Query range
	queryRange := promApiV1.Range{
		Start: start,
		End:   end,
		Step:  resolutionStep,
	}
	val, _, err := api.QueryRange(context.Background(), pq, queryRange)
	if err != nil {
		return nil, err
	}
	matrix := val.(promModel.Matrix)

	return matrix, nil
}
