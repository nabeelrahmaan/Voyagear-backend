package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WishlistItem struct {
	ID         uuid.UUID
	WishlistID uuid.UUID
	ProductID  uuid.UUID

	Product Product `gorm:"foreignKey:ProductID"`

	CreatedAt time.Time
}

func (w *WishlistItem) BeforeCreate(tx *gorm.DB) error {
	w.ID = uuid.New()
	return nil
}