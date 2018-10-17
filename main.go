package main

import (
	"github.com/gin-gonic/gin"
	"github.com/hekike/outlier-istio/src/models"
)

const addr = "http://192.168.99.100:30900"

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.Status(200)
	})
	r.GET("/services", func(c *gin.Context) {
		services, err := models.GetServices(addr)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.JSON(200, gin.H{
			"services": services,
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
