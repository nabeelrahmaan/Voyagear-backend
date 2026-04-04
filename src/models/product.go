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
	Variants      []Variant `gorm:"foreignKey:ProductID;references:ID"`

	CreatedAt time.Time 
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	p.ID = uuid.New()
	return nil
}

type ProductResponse struct {
	ID            uuid.UUID          `json:"id"`
    Name          string             `json:"name"`
    Category      string             `json:"category"`
    ImageURL      string             `json:"image_url"`
    Description   string             `json:"description"`
    Price         int                `json:"price"`
    OriginalPrice int                `json:"original_price"`
    Stock         int                `json:"stock"`
    IsActive      bool               `json:"is_active"`
    Variants      []VariantResponse  `json:"variants" gorm:"-"`
    IsWishlisted  bool               `json:"is_wishlisted" gorm:"-"`
    CreatedAt     time.Time          `json:"created_at"`
    UpdatedAt     time.Time          `json:"updated_at"`
    DeletedAt     gorm.DeletedAt     `json:"deleted_at"`
}
