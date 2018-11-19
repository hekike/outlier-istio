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

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/hekike/outlier-istio/src/models"
)

// APIResponseWorkloads struct.
type APIResponseWorkloads struct {
	Workloads []models.Workload `json:"workloads"`
}

// Setup router
func Setup(promAddr string, webDistPath string) *gin.Engine {
	router := gin.Default()
	apiRouter := router.Group("/api/v1")

	router.Use(static.Serve("/", static.LocalFile(webDistPath, false)))

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
		c.String(http.StatusOK, "ok")
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
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
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
		c.JSON(http.StatusOK, response)
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
	// 	- name: historical
	// 	  in: query
	// 	  schema:
	// 	    type: int
	// 	  description: Historical data in minutes
	// 	- name: statusStep
	// 	  in: query
	// 	  schema:
	// 	    type: int
	// 	  description: Status steps in minutes
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
		Start      time.Time `form:"start" time_format:"2006-01-02T15:04:05Z07:00"`
		End        time.Time `form:"end" time_format:"2006-01-02T15:04:05Z07:00"`
		Historical int       `form:"historical"`
		StatusStep int       `form:"statusStep"`
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
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Parameter defaults
		if status.End.IsZero() {
			status.End = time.Now()
		}
		if status.Start.IsZero() {
			status.Start = status.End.Add(-time.Hour)
		}
		if status.Historical == 0 {
			status.Historical = 15
		}
		if status.StatusStep == 0 {
			status.StatusStep = 5
		}

		historical := time.Duration(status.Historical) * time.Minute
		statusStep := time.Duration(status.StatusStep) * time.Minute

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
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Response
		c.JSON(http.StatusOK, workload)
	})

	router.Use(func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	return router
}
