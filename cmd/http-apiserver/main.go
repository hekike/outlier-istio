package main

import (
	"os"

	"github.com/hekike/outlier-istio/pkg/router"
)

func main() {
	promAddr := os.Getenv("PROMETHEUS_HOST")
	webDistPath := os.Getenv("WEB_DIST_PATH")

	r := router.Setup(promAddr, webDistPath)
	r.Run() // listen and serve on 0.0.0.0:8080
}
