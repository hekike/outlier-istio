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
	"net/http"
	"sort"
	"time"

	"github.com/gin-contrib/cors"
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
	apiRouter.Use(cors.Default())

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
		// Get data
		workloadsMap, err := models.GetWorkloads(promAddr)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{
				"error": err,
			})
			return
		}

		// Sort map keys
		keys := make([]string, 0, len(workloadsMap))
		for k := range workloadsMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Convert map to slices
		workloads := make([]models.Workload, 0, len(workloadsMap))
		for _, k := range keys {
			workloads = append(workloads, workloadsMap[k])
		}

		// Response
		response := APIResponseWorkloads{Workloads: workloads}
		c.JSON(200, response)
	})

	// swagger:route GET /api/v1/workloads/{name}/status workload getWorkloadStatusByName
	// ---
	// summary: Returns with destination workloads
	// description: Returns with an array of services.
	// parameters:
	// 	- name: name
	// 	  in: path
	// 	  schema:
	// 	    type: string
	//	  description: Name of the workload
	// 	- name: start
	// 	  in: query
	// 	  schema:
	// 	    type: string
	// 	    format: date
	// 	  description: The start date for the report.
	// 	- name: end
	// 	  in: query
	// 	  schema:
	// 	    type: string
	// 	    format: date
	// 	  description: The end date for the report.
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
	type Status struct {
		Start time.Time `form:"start" time_format:"2006-01-02T15:04:05Z07:00"`
		End   time.Time `form:"end" time_format:"2006-01-02T15:04:05Z07:00"`
	}

	apiRouter.GET("/workloads/:name/status", func(c *gin.Context) {
		name := c.Param("name")

		// Validation
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Workload name cannot be empty",
			})
		}

		// Bind query string parameters
		var status Status
		err := c.ShouldBindQuery(&status)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		// Parameter defaults
		if status.End.IsZero() {
			status.End = time.Now()
		}
		if status.Start.IsZero() {
			status.Start = status.End.Add(-time.Hour)
		}
		historical := 15 * time.Minute
		statusStep := 5 * time.Minute

		// Get data
		workload, err := models.GetWorkloadStatusByName(
			promAddr,
			name,
			status.Start,
			status.End,
			historical,
			statusStep,
		)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		// Response
		c.JSON(200, workload)
	})

	return router
}
