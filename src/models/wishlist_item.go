package models

import (
	"time"

	"github.com/google/uuid"
)

type WishlistItem struct {
	ID         uuid.UUID
	WishlistID uuid.UUID
	ProductID  uuid.UUID

	Product Product `gorm:"foreignKey:ProductID"`

	CreatedAt time.Time
}
