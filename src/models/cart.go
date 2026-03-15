package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Cart struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;uniqueIndex"`
	User      User       `gorm:"foreignKey:UserID"`

	Items     []CartItem `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE"`
	
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (c *Cart) BeforeCreate(tx *gorm.DB) error {
	c.ID = uuid.New()
	return nil
}
