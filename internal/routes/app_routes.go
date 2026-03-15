package routes

import (
	"voyagear/src/controller"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(auth *controller.AuthController) *gin.Engine {
	r := gin.Default()

	authGroup := r.Group("/auth")
	authGroup.POST("/signup", auth.Signup)
	authGroup.POST("/login", auth.Login)
	authGroup.POST("/verify-otp", auth.VerifyOTP)
	authGroup.POST("/forget-password", auth.ForgotPassword)
	authGroup.POST("/reset-password", auth.ResetPassword)
	r.POST("/refresh", auth.RefreshToken)

	userGroup := r.Group("/user")
	userGroup.GET("/profile", auth.GetProfile)
	userGroup.PUT("/profile", auth.UpdateProfile)

	return r
}
