package fixtures

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// PrometheusResponseStub creates a mock server
func PrometheusResponseStub(t *testing.T, filename map[string]string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		json, err := ioutil.ReadFile(filename[r.FormValue("query")])
		if err != nil {
			t.Error(err)
		}
		resp := string(json)
		w.Write([]byte(resp))
	}))
	return server
}

// HTTPRequest creates a request to an HTTP server
func HTTPRequest(t *testing.T, url string) (res *http.Response, body []byte) {
	client := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Error(err)
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		t.Error(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		t.Error(readErr)
	}

	return res, body
}
