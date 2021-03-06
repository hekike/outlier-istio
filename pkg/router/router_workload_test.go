package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hekike/outlier-istio/pkg/models"
	"github.com/hekike/outlier-istio/pkg/prometheus"
	"github.com/hekike/outlier-istio/test/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestApiGetWorkloads(t *testing.T) {
	mockServer := fixtures.PrometheusResponseStub(t, map[string]string{
		prometheus.GetRequestsTotalByWorkloadsQuery(): "../../test/mock/prom_workload_request_totals.json",
	})
	defer mockServer.Close()

	// router
	testRouter := Setup(mockServer.URL, "./web-dist")
	server := httptest.NewServer(testRouter)

	// call api
	workloadsURL := server.URL + "/api/v1/workloads"
	res, body := fixtures.HTTPRequest(t, workloadsURL)

	workloadsResponse := APIResponseWorkloads{}
	jsonErr := json.Unmarshal(body, &workloadsResponse)
	if jsonErr != nil {
		panic(jsonErr)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.ElementsMatch(
		t,
		getWorkloadsResponseMock(),
		workloadsResponse.Workloads,
	)
}

func getWorkloadsResponseMock() []models.Workload {
	unknown := models.Workload{}
	unknown.Name = "unknown"
	unknown.App = "unknown"
	unknown.Sources = make([]models.Workload, 0)

	productpage := models.Workload{}
	productpage.Name = "productpage-v1"
	productpage.App = "productpage"

	reviews := models.Workload{}
	reviews.Name = "reviews-v3"
	reviews.App = "reviews"

	ratings := models.Workload{}
	ratings.Name = "ratings-v1"
	ratings.App = "ratings"
	ratings.Destinations = make([]models.Workload, 0)

	unknown.AddDestination(
		models.Workload{Name: "productpage-v1", App: "productpage"},
	)

	productpage.AddSource(models.Workload{Name: "unknown", App: "unknown"})
	productpage.AddDestination(
		models.Workload{Name: "reviews-v3", App: "reviews"},
	)

	reviews.AddSource(
		models.Workload{Name: "productpage-v1", App: "productpage"},
	)
	reviews.AddDestination(
		models.Workload{Name: "ratings-v1", App: "ratings"},
	)

	ratings.AddSource(models.Workload{Name: "reviews-v3", App: "reviews"})

	workloads := []models.Workload{unknown, reviews, ratings, productpage}

	return workloads
}
