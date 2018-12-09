package router

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/hekike/outlier-istio/pkg/models"
)

// APIResponseWorkloads struct.
type APIResponseWorkloads struct {
	Workloads []models.Workload `json:"workloads"`
}

// RegisterRouteGroupWorkload register route
func RegisterRouteGroupWorkload(promAddr string, r *gin.RouterGroup) {
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
	r.GET("/workloads", func(c *gin.Context) {
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
}
