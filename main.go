package main

import (
	"github.com/hekike/outlier-istio/src"
)

const promAddr = "http://192.168.99.100:30900"

func main() {
	router := router.Setup(promAddr)
	router.Run() // listen and serve on 0.0.0.0:8080
}
