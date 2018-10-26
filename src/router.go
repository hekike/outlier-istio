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

// Package router outlier-istio
package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hekike/outlier-istio/src/models"
)

// APIResponseWorkloads struct.
type APIResponseWorkloads struct {
	Workloads []models.Workload `json:"workloads"`
}

// Setup router
func Setup(promAddr string) *gin.Engine {
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

	// swagger:route GET /api/v1/workloads workload getWorkloads
	// ---
	// summary: Returns with destination workloads
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
	apiRouter.GET("/workloads", func(c *gin.Context) {
		workloadsMap, err := models.GetWorkloads(promAddr)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		// Convert slice to array
		workloads := make([]models.Workload, 0, len(workloadsMap))
		for _, workload := range workloadsMap {
			workloads = append(workloads, workload)
		}

		response := APIResponseWorkloads{Workloads: workloads}

		c.JSON(200, response)
	})

	// swagger:route GET /api/v1/workloads workload getWorkloads
	// ---
	// summary: Returns with destination workloads
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
	apiRouter.GET("/workloads/:name/status", func(c *gin.Context) {
		name := c.Param("name")
		end := time.Now()
		start := end.Add(-time.Hour)
		historical := 60 * time.Minute
		statusStep := 5 * time.Minute

		workload, err := models.GetWorkloadStatusByName(
			promAddr,
			name,
			start,
			end,
			historical,
			statusStep,
		)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.JSON(200, workload)
	})

	return router
}
