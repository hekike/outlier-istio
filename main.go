// Outlier Istio
//
// This documentation describes the https://github.com/hekike/outlier-istio API.
//
//     Schemes: http
//     BasePath: /api/v1
//     Version: 1.0.0
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
// swagger:meta
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/hekike/outlier-istio/src/models"
)

const addr = "http://192.168.99.100:30900"

func main() {
	router := gin.Default()
	apiRouter := router.Group("/api/v1")

	// swagger:route GET /ping operation ping
	// ---
	// summary: Returns with server healthcheck as a status code.
	// description: Returns 200 for healthy and 500 for unhealtyh instance.
	// servers:
	//	url: /
	// produces:
	// 	- application/json
	// schemes:
	// 	- http
	// responses:
	// 	default:
	//		description: Unexpected error
	// 	200:
	//		type: string
	//		description: OK
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// swagger:route GET /api/v1/services service getServices
	// ---
	// summary: Returns with services and downstream relations.
	// description: Returns with an array of services.
	// produces:
	// 	- application/json
	// schemes:
	// 	- http
	// responses:
	// 	default:
	//		description: Unexpected error
	// 	200:
	//		type: string
	//		description: TODO
	apiRouter.GET("/services", func(c *gin.Context) {
		services, err := models.GetServices(addr)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.JSON(200, gin.H{
			"services": services,
		})
	})

	router.Run() // listen and serve on 0.0.0.0:8080
}
