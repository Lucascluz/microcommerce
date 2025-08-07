package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lucas/shared/utils"
)

func main() {
	port := utils.GetEnvOrDefault("PORT", "8083")

	router := gin.Default()

	// Api gateway health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "user-service",
		})
	})

	log.Printf("User service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
