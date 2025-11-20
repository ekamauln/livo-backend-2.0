package middleware

import (
	"net/http"
	"strings"

	"livo-backend-2.0/config"
	"livo-backend-2.0/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Header authorization dibutuhkan", "header authorization tidak ditemukan")
			c.Abort()
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Format header authorization tidak valid", "format bearer token tidak valid")
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(bearerToken[1], cfg.JWTSecret)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Token tidak valid", err.Error())
			c.Abort()
			return
		}

		// Set user claims in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}
