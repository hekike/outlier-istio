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

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

// Setup router
func Setup(promAddr string, webDistPath string) *gin.Engine {
	router := gin.Default()
	apiRouter := router.Group("/api/v1")

	router.Use(static.Serve("/", static.LocalFile(webDistPath, false)))

	// Routes
	RegisterRouteGroupPing(router)

	// API routes
	RegisterRouteGroupWorkload(promAddr, apiRouter)
	RegisterRouteGroupWorkloadStatus(promAddr, apiRouter)

	router.Use(func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
	})

	return router
}
