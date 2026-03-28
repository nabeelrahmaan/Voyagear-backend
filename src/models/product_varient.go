package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Variant struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID uuid.UUID `gorm:"type:uuid;uniquIndex:idx_product_size"`

	Product Product `gorm:"foreignKey;ProductID;references:ID"`

	Size string `gorm:"uniquIndex:idx_product_size"`

	Quantity  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *Variant) BeforeCreate(tx *gorm.DB) error {
	p.ID = uuid.New()
	return nil
}
