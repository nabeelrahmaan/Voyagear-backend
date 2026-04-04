package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Variant struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_product_size"`

	Product Product `gorm:"foreignKey:ProductID;references:ID" json:"-"`

	Size string `gorm:"uniqueIndex:idx_product_size"`

	Quantity  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *Variant) BeforeCreate(tx *gorm.DB) error {
	p.ID = uuid.New()
	return nil
}


type VariantResponse struct {
	ID        uuid.UUID `json:"id"`
	Size      string    `json:"size"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}