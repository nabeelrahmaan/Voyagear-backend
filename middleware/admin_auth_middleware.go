package middleware

import (
	"voyagear/src/models"
	"voyagear/src/repository"
	"voyagear/utils/constant"

	"github.com/gin-gonic/gin"
)

func AdminAuthMiddleware(repo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {

		userID, exist := c.Get("user_id")
		if !exist || userID == "" {
			c.JSON(constant.UNAUTHORIZED, gin.H{"error": "User not found"})
			c.Abort()
			return 
		}

		var user models.User
		if err := repo.FindById(&user, userID); err != nil {
			c.JSON(constant.UNAUTHORIZED, gin.H{"error":"User not found"})
			c.Abort()
			return 
		}

		if user.IsBlocked {
			c.JSON(constant.FORBIDDEN, gin.H{"error":"Your account has been blocked"})
			c.Abort()
			return 
		}

		if user.Role != "admin" {
			c.JSON(constant.FORBIDDEN, gin.H{"error":"Admin access required"})
			c.Abort()
			return 
		}

		c.Next()
	}
}