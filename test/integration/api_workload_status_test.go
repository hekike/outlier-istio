package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hekike/outlier-istio/src"
	"github.com/hekike/outlier-istio/src/models"
	"github.com/hekike/outlier-istio/test/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestApiGetWorkloadStatus(t *testing.T) {
	mockServer := fixtures.PrometheusResponseStub(t, "../data/prom_workload_source_request_duration_95th_10m_1m.json")
	defer mockServer.Close()

	workloadName := "productpage-v1"

	// router
	testRouter := router.Setup(mockServer.URL)
	server := httptest.NewServer(testRouter)

	// call api
	workloadsURL := server.URL + "/api/v1/workloads/" + workloadName + "/status?end=2018-10-27T15:00:00Z"
	res, body := fixtures.HTTPRequest(t, workloadsURL)

	workloadsResponse := models.Workload{}
	jsonErr := json.Unmarshal(body, &workloadsResponse)
	if jsonErr != nil {
		panic(jsonErr)
	}

	expected := models.Workload{}
	expected.Name = workloadName

	assert.Equal(t, http.StatusOK, res.StatusCode)

	detailsV1 := workloadsResponse.Destinations[0]
	assert.Equal(t, "details-v1", detailsV1.Name)
	assert.Equal(t, "details", detailsV1.App)

	statuses := make([]string, len(detailsV1.Statuses))

	for i, status := range detailsV1.Statuses {
		statuses[i] = status.Status
	}

	assert.Equal(t, []string{
		"high", "ok", "ok", "ok",
		"ok", "high", "ok", "high",
		"ok", "high", "ok", "ok",
		"high", "high", "high", "ok",
	}, statuses)
}
