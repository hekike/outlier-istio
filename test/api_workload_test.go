package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/hekike/outlier-istio/src"
)

const promAddr = "http://192.168.99.100:30900"

func TestItem(t *testing.T) {
	// router
	testRouter := router.Setup(db)
	server := httptest.NewServer(testRouter)
	e := httpexpect.New(t, server.URL)

	// test ping
	e.GET("/ping").
		Expect().
		Status(http.StatusOK)
	})

	// test invalid body
	e.GET("/api/workloads").
		Expect().
		Status(http.StatusBadRequest).JSON().Object().Equal(map[string]interface{}{
		"error":      "Bad Request",
		"message":    "Invalid ObjectId",
		"statusCode": 400,
	})
}
