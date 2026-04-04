package main

import (
	"fmt"
	"time"
	"voyagear/config"
	"voyagear/internal/cache"
	"voyagear/internal/routes"
	"voyagear/migration"
	"voyagear/src/controller"
	"voyagear/src/database"
	"voyagear/src/repository"
	"voyagear/src/services"
	"voyagear/utils/email"
	"voyagear/utils/jwt"
	"voyagear/utils/logger"
	"voyagear/utils/razorpay"
	"voyagear/utils/validation"

	"github.com/gin-gonic/gin"
)

func main() {

	cfg, err := config.LoadConfig("app.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v", err)
		return
	}

	db := database.SetupDatabase(cfg)

	migration.Migrate(db)

	// Bind generic structured JSON parsers to use the custom validation templates unconditionally
	validation.InitValidation()

	// Intercept boot sequence to spin up explicit logrus formatting and directories natively
	appLogger := logger.InitLogger()
	gin.DefaultWriter = appLogger.Out

	r := gin.Default()

	repo := repository.SetupRepo(db)

	redis := cache.NewRedis()

	email.Init(cfg.SMTP)

	jwtManager := jwt.GenerateJWT(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		time.Minute*time.Duration(cfg.JWT.AccessTTLMinutes),
		time.Hour*time.Duration(cfg.JWT.RefreshTTLHours),
		time.Hour*time.Duration(cfg.JWT.MaxSessionHours),
	)

	authService := services.CreateAuthService(repo, redis, jwtManager)
	authController := controller.NewAuthController(authService)

	productService := services.SetupProductService(repo)
	productController := controller.SetupProductController(productService)

	cartService := services.SetupCart(repo)
	cartController := controller.SetupCartController(cartService)

	addressService := services.SetupAddressService(repo)
	addressController := controller.SetupAddressController(addressService)

	wishlistService := services.SetupWishlist(repo)
	wishlistController := controller.SetupWishlistController(wishlistService)

	// Initialize Gateway Settings
	rzpClient := razorpay.NewRazorpayClient(&cfg.Razorpay)

	orderService := services.SetupOrderService(repo, rzpClient)
	orderController := controller.SetupOrderController(orderService)

	paymentService := services.SetupPaymentService(repo, rzpClient)
	paymentController := controller.SetupPaymentController(paymentService)

	routes.SetupRoutes(
		r,
		authController,
		productController,
		cartController,
		wishlistController,
		orderController,
		paymentController,
		addressController,
		jwtManager,
		repo,
	)

	r.Run(":" + cfg.Server.Port)
}
