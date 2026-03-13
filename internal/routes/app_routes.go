package routes

import "github.com/gin-gonic/gin"

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	auth := r.Group("/auth")
	{
		auth.POST("/signup")
		auth.POST("/login")
		auth.POST("/send-otp")
		auth.POST("/verify-otp")
		auth.POST("/forget-password")
		auth.POST("/reset-password")
	}
	return r
}