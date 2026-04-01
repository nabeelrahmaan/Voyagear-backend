package routes

import (
	"voyagear/middlewear"
	"voyagear/src/controller"
	"voyagear/src/repository"
	"voyagear/utils/jwt"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	auth *controller.AuthController,
	product *controller.ProductController,
	cart *controller.CartController,
	wishlist *controller.WishlistController,
	order *controller.OrderController,
	jwtManager *jwt.JWTmanger,
	repo *repository.Repository,
) {
	
	r.GET("/",auth.Test)
	// ========== Auth rotes (public) ==========
	authGroup := r.Group("/auth")
	authGroup.POST("/signup", auth.Signup)
	authGroup.POST("/login", auth.Login)
	authGroup.POST("/verify-otp", auth.VerifyOTP)
	authGroup.POST("/forget-password", auth.ForgotPassword)
	authGroup.POST("/reset-password", auth.ResetPassword)
	authGroup.POST("/logout", auth.Logout)
	r.POST("/refresh", auth.RefreshToken)

	// ========== User routes (protected) ==========
	userGroup := r.Group("/user", middlewear.AuthMiddleware(jwtManager))

	// Profile
	userGroup.GET("/profile", auth.GetProfile)
	userGroup.PUT("/profile", auth.UpdateProfile)

	// Cart
	cartGroup := userGroup.Group("/cart")
	cartGroup.POST("/", cart.AddToCart)
	cartGroup.GET("/", cart.GetCart)
	cartGroup.PUT("/:id", cart.UpdateCartItem)
	cartGroup.DELETE("/:id", cart.RemoveCartItem)

	// Wishlist
	wishlistGroup := userGroup.Group("/wishlist")
	wishlistGroup.POST("/", wishlist.AddToWishlist)
	wishlistGroup.GET("/", wishlist.GetWiahlist)
	wishlistGroup.DELETE("/:id", wishlist.RemoveFromWishlist)

	// Order
	orderGroup := userGroup.Group("/order")
	orderGroup.GET("/", order.GetUserOrders)
	orderGroup.POST("/", order.PlaceOrder)
	orderGroup.GET("/:id", order.GetOrderDetails)
	orderGroup.PUT("/:id/cancel", order.UpdateOrderStatusUser)
	orderGroup.DELETE("/:id", order.DeleteOrder)


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

	// Orders
	adminGroup.GET("/orders", order.GetAllOrders)
	adminGroup.PUT("/orders/:id", order.UpdateOrderStatusAdmin)

}
