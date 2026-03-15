package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserAddress struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    string `gorm:"type:uuid"`
	Line1     string
	Line2     string
	City      string
	State     string
	Zipcode   string
	IsDefault bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u *UserAddress) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}