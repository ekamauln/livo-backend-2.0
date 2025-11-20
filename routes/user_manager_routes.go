package routes

import (
	"livo-backend-2.0/config"
	"livo-backend-2.0/controllers"
	"livo-backend-2.0/middleware"

	"github.com/gin-gonic/gin"
)

// SetupUserManagerRoutes configures user manager-related routes
func SetupUserManagerRoutes(api *gin.RouterGroup, cfg *config.Config, userManagerController *controllers.UserManagerController) {
	// User manager routes (authenticated + role-based)
	userManager := api.Group("/user-manager")
	userManager.Use(middleware.AuthMiddleware(cfg))
	{
		// Get all users - public to all authenticated users (no role restriction)
		userManager.GET("/users", userManagerController.GetUsers)
		userManager.GET("/users/:id", userManagerController.GetUser)

		// Roles management - public to all authenticated users (no role restriction)
		userManager.GET("/roles", userManagerController.GetRoles)

		// User management (coordinator)
		users := userManager.Group("/users")
		users.Use(middleware.RequireUserManagementRoles())
		{
			users.PUT("/:id/status", userManagerController.UpdateUserStatus)     // Update user status (active/inactive)
			users.PUT("/:id/password", userManagerController.UpdateUserPassword) // Update user password
			users.PUT("/:id/profile", userManagerController.UpdateUserProfile)   // Update user profile
			users.POST("", userManagerController.CreateUser)                     // Create new user
			users.DELETE("/:id", userManagerController.DeleteUser)               // Delete user
		}

		// Assign or remove roles to/from a user
		// Role assignment (coordinator)
		roleAssignment := userManager.Group("/users/:id/roles")
		roleAssignment.Use(middleware.RequireUserManagementRoles())
		{
			roleAssignment.POST("", userManagerController.AssignRole)   // Assign role to user
			roleAssignment.DELETE("", userManagerController.RemoveRole) // Remove role from user
		}
	}
}
