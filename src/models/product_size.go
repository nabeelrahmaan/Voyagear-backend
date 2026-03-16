package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductSize struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID uuid.UUID `gorm:"type:uuid;unique"`
	Size      string
	Quantity  int
	CreatedAt time.Time 
	UpdatedAt time.Time 
}

func (p *ProductSize) BeforeCreate(tx *gorm.DB) error {
	p.ID = uuid.New()
	return nil
}
