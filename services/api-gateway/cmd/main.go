package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lucas/api-gateway/handlers"
	"github.com/lucas/shared/utils"
)

func main() {
	port := utils.GetEnvOrDefault("PORT", "8080")

	router := gin.Default()

	// Api gateway health check
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "api-gateway",
		})
	})

	// Initialize gateway handler
	gatewayHandler := handlers.NewGatewayHandler()

	// Service routes
	api := router.Group("/api/v1")
	{
		api.GET("/services/health", gatewayHandler.PingServices)
	}

	log.Printf("API Gateway starting on port %s", port)
	// Starting server with updated handlers
	log.Fatal(http.ListenAndServe(":"+port, router))
}
