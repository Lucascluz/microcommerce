package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/lucas/api-gateway/handlers"
	"github.com/lucas/api-gateway/middleware"
)

func SetupRoutes(router *gin.Engine) {

	// Initialize handlers
	gatewayHandler := handlers.NewGatewayHandler()
	userHandler := handlers.NewUserHandler()

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware()

	// Gateway health check
	router.GET("/", gatewayHandler.HealthCheck)

	api := router.Group("/api/v1")
	{
		// Gateway routes
		api.GET("services/health", gatewayHandler.PingServices)

		// Public user routes (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		// Protected user routes (auth required)
		users := api.Group("/users")
		users.Use(authMiddleware.RequireAuth())
		{
			auth.POST("/logout", authMiddleware.RequireAuth(), userHandler.Logout)
			users.GET("/profile", userHandler.GetProfile)
		}
	}

}
