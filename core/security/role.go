package security

import (
	"FinanceGolang/core/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "user not found",
			})
			c.Abort()
			return
		}

		userModel := user.(*domain.User)
		for _, role := range roles {
			if userModel.HasRole(role) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "insufficient permissions",
		})
		c.Abort()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return RoleMiddleware(domain.RoleAdmin)
}
