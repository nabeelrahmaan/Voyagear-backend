package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Variants struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID uuid.UUID `gorm:"type:uuid;"`
	Size      string
	Quantity  int
	CreatedAt time.Time 
	UpdatedAt time.Time 
}

func (p *Variants) BeforeCreate(tx *gorm.DB) error {
	p.ID = uuid.New()
	return nil
}
