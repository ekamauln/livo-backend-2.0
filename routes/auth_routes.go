package routes

import (
	"livo-backend-2.0/config"
	"livo-backend-2.0/controllers"
	"livo-backend-2.0/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes configures authentication routes
func SetupAuthRoutes(api *gin.RouterGroup, cfg *config.Config, authController *controllers.AuthController) {
	// Auth routes (public)
	auth := api.Group("/auth")
	{
		// Public auth routes
		auth.POST("/register", authController.Register)                             // User registration
		auth.POST("/login", authController.Login)                                   // User login
		auth.POST("/refresh", authController.RefreshToken)                          // Refresh access token
		auth.POST("/logout", middleware.AuthMiddleware(cfg), authController.Logout) // User logout
	}
}
