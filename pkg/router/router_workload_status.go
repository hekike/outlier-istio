package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hekike/outlier-istio/pkg/models"
)

// RegisterRouteGroupWorkloadStatus register route
func RegisterRouteGroupWorkloadStatus(promAddr string, r *gin.RouterGroup) {
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

	r.GET("/workloads/:name/status", func(c *gin.Context) {
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
}
