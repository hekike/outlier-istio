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
	// router
	testRouter := router.Setup("http://localhost", "./web-dist")
	server := httptest.NewServer(testRouter)

	// test ping
	testURL := server.URL + "/ping"
	res, _ := fixtures.HTTPRequest(t, testURL)

	assert.Equal(t, http.StatusOK, res.StatusCode)

}
