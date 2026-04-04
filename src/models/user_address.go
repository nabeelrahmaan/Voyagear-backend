package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserAddress struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid" json:"user_id" validate:"required"` 
	User      User      `gorm:"foreignKey:UserID"`
	Line1     string	`json:"line_1"`
	Line2     string	`json:"line_2"`
	City      string	`json:"city" validate:"required"`
	State     string	`json:"state" validate:"required"`
	Zipcode   string    `json:"zipcode" validate:"required"`
	IsDefault bool		`gorm:"default:true" json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *UserAddress) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}
