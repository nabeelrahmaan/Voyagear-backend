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

	"github.com/gin-gonic/gin"
)

func main() {

	cfg, err := config.LoadConfig("app.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v", err)
		return
	}

	db := database.SetupDatbase(cfg)

	migration.Migrate(db)

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

	

		routes.SetupRoutes(
			r,
			authController,
			productController,
			jwtManager,
			repo,
		)
	
		r.Run(":" + cfg.Server.Port)
}