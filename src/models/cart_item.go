package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CartItem struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	CartID    uuid.UUID `gorm:"type:uuid;index"`
	ProductID uuid.UUID `gorm:"type:uuid;index"`
	VariantID uuid.UUID `gorm:"type:uuid;index"`

	Quantity int

	Product Product `gorm:"foreignKey:ProductID"`
	Variant Variant `gorm:"foreignKey:VariantID"`

	CreatedAt time.Time
}

func (c *CartItem) BeforeCreate(tx *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}
