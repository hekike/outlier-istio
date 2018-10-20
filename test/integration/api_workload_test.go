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

func TestApiGetWorkloads(t *testing.T) {
	mockServer := fixtures.PrometheusResponseStub(t, "../data/prom_workloads.json")
	defer mockServer.Close()

	// router
	testRouter := router.Setup(mockServer.URL)
	server := httptest.NewServer(testRouter)

	// call api
	workloadsURL := server.URL + "/api/v1/workloads"
	res, body := fixtures.HTTPRequest(t, workloadsURL)

	workloadsResponse := router.APIResponseWorkloads{}
	jsonErr := json.Unmarshal(body, &workloadsResponse)
	if jsonErr != nil {
		panic(jsonErr)
	}

	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.ElementsMatch(t, workloadsResponse.Workloads, getWorkloadsResponseMock())
}

func getWorkloadsResponseMock() []models.Workload {
	unknown := models.Workload{}
	unknown.Name = "unknown"
	unknown.App = "unknown"
	unknown.Sources = make([]models.WorkloadItem, 0)

	productpage := models.Workload{}
	productpage.Name = "productpage-v1"
	productpage.App = "productpage"

	reviews := models.Workload{}
	reviews.Name = "reviews-v3"
	reviews.App = "reviews"

	ratings := models.Workload{}
	ratings.Name = "ratings-v1"
	ratings.App = "ratings"
	ratings.Destinations = make([]models.WorkloadItem, 0)

	unknown.AddDestination(models.WorkloadItem{Name: "productpage-v1", App: "productpage"})

	productpage.AddSource(models.WorkloadItem{Name: "unknown", App: "unknown"})
	productpage.AddDestination(models.WorkloadItem{Name: "reviews-v3", App: "reviews"})

	reviews.AddSource(models.WorkloadItem{Name: "productpage-v1", App: "productpage"})
	reviews.AddDestination(models.WorkloadItem{Name: "ratings-v1", App: "ratings"})

	ratings.AddSource(models.WorkloadItem{Name: "reviews-v3", App: "reviews"})

	workloads := []models.Workload{unknown, reviews, ratings, productpage}

	return workloads
}
