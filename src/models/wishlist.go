package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Wishlist struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`

	Items     []WishlistItem

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (w *Wishlist) BeforeCreate(tx *gorm.DB) error {
	w.ID = uuid.New()
	return nil
}
