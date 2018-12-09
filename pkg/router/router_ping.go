package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRouteGroupPing register route
func RegisterRouteGroupPing(r *gin.Engine) {
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
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
}
