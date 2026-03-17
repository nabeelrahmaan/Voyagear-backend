package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	Name        string `gorm:"not null"`
	Category    string `gorm:"not null"`
	ImageURL    string
	Description string

	Price         int `gorm:"not null"`
	OriginalPrice int
	Stock         int
	IsActive      bool 
	Variants      []Variants `gorm:"foreignKey:ProductID;references:ID"`

	CreatedAt time.Time 
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	p.ID = uuid.New()
	return nil
}
