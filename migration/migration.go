package migration

import (
	"fmt"
	"voyagear/src/models"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	models := []interface{}{
		&models.User{},
		&models.Product{},
		&models.Variant{},
		&models.RefreshToken{},
		&models.Cart{},
		&models.CartItem{},
		&models.Wishlist{},
		&models.WishlistItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.UserAddress{},
	}

	if err := db.AutoMigrate(models...); err != nil {
		fmt.Println("Failed to automigrate models")
		return
	}
}

