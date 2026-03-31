package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	Total     int
	Status    string
	Items     []OrderItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	o.ID = uuid.New()
	return nil
}
