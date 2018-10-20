package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hekike/outlier-istio/src"
	"github.com/hekike/outlier-istio/test/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestApiGetPing(t *testing.T) {
	mockServer := fixtures.PrometheusResponseStub(t, "prom_workloads.json")
	defer mockServer.Close()

	// router
	testRouter := router.Setup(mockServer.URL)
	server := httptest.NewServer(testRouter)

	// test ping
	workloadsURL := server.URL + "/ping"
	res, _ := fixtures.HTTPRequest(t, workloadsURL)

	assert.Equal(t, res.StatusCode, http.StatusOK)

}
