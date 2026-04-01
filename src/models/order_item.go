package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	OrderID   uuid.UUID `gorm:"type:uuid"`
	ProductID uuid.UUID `gorm:"type:uuid"`
	VariantID uuid.UUID `gorm:"type:uuid"`

	Quantity int
	Price    int

	Product   Product `gorm:"foreignKey:ProductID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) error {
	oi.ID = uuid.New()
	return nil
}
