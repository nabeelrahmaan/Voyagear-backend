package middleware

import (
	"strings"
	"voyagear/utils/constant"
	"voyagear/utils/jwt"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtManager *jwt.JWTmanger) gin.HandlerFunc {
	return func (c *gin.Context) {
		
		token, err := c.Cookie("access_token")
		if err != nil || token == "" {

			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(constant.UNAUTHORIZED, gin.H{"error": "Authorization header missing"})
				c.Abort()
				return 
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(constant.UNAUTHORIZED, gin.H{"error": "Invalid authorization format"})
				c.Abort()
				return 
			}

			token = parts[1]
		}

		claims, err := jwtManager.ValidateAccess(token)
		if err != nil {
			c.JSON(constant.UNAUTHORIZED, gin.H{"error":"Invalid access token"})
			c.Abort()
			return 
		}

		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			c.JSON(constant.UNAUTHORIZED, gin.H{"error":"Invalid token claims"})
			c.Abort()
			return 
		}

		role := claims["role"].(string)
		
		c.Set("user_id", userID)
		if role != "" {
			c.Set("role", role)
		}

		c.Next()
	}
}