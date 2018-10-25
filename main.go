package main

import (
	"github.com/hekike/outlier-istio/src"
)

const promAddr = "http://localhost:9090"

func main() {
	router := router.Setup(promAddr)
	router.Run() // listen and serve on 0.0.0.0:8080
}
