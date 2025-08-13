package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "github.com/lucas/api-gateway/routes"
    "github.com/lucas/shared/utils"
)

func main() {
    port := utils.GetEnvOrDefault("PORT", "8080")

    router := gin.Default()

    // Setup all routes
    routes.SetupRoutes(router)

    log.Printf("API Gateway starting on port %s", port)
    log.Fatal(router.Run(":" + port))
}