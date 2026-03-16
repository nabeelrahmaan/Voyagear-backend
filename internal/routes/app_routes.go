package routes

import (
	"voyagear/middlewear"
	"voyagear/src/controller"
	"voyagear/src/repository"
	"voyagear/utils/jwt"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	auth *controller.AuthController,
	product *controller.ProductController,
	jwtManager *jwt.JWTmanger,
	repo *repository.Repository,
) *gin.Engine {
	r := gin.Default()

	// ========== Auth rotes (public) ==========
	authGroup := r.Group("/auth")
	authGroup.POST("/signup", auth.Signup)
	authGroup.POST("/login", auth.Login)
	authGroup.POST("/verify-otp", auth.VerifyOTP)
	authGroup.POST("/forget-password", auth.ForgotPassword)
	authGroup.POST("/reset-password", auth.ResetPassword)
	r.POST("/refresh", auth.RefreshToken)

	// ========== User routes (protected) ==========
	userGroup := r.Group("/user", middlewear.AuthMiddleware(jwtManager))

	// Profile
	userGroup.GET("/profile", auth.GetProfile)
	userGroup.PUT("/profile", auth.UpdateProfile)


	// ========== Public product routes ==========
	r.GET("/products", product.GetAllProducts)
	r.GET("/products/search", product.SearchProduct)
	r.GET("/product/:id", product.GetProductById)


	// ========== Admin routes (protected) ==========
	adminGroup := r.Group("/admin", middlewear.AuthMiddleware(jwtManager), middlewear.AdminAuthMiddleware(*repo))

	// Users
	adminGroup.GET("/users", auth.GetAllUsers)
	adminGroup.PUT("/users/:id/block", auth.ToggleISBlock)
	adminGroup.DELETE("/users/:id", auth.DeleteUserById)

	// Products
	adminGroup.POST("/products", product.CreateProduct)
	adminGroup.PATCH("/products/:id", product.UpdateProduct)
	adminGroup.DELETE("/product/:id", product.DeleteProduct)

	return r
}
