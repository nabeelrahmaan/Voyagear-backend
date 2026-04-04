package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID            uuid.UUID `gorm:"type:uuid"`
	AddressID         uuid.UUID `gorm:"type:uuid"`
	Address           UserAddress `gorm:"foreignKey:AddressID"`

	Total             int
	Status            string
	PaymentMethod	  string
	PaymentStatus 	  string `gorm:"default:pending"`

	RazorpayOrderID   string `gorm:"type:varchar(100)" json:"razorpay_order_id,omitempty"`
	RazorpayPaymentID string `gorm:"type:varchar(100)" json:"razorpay_payment_id,omitempty"`

	Items             []OrderItem
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	o.ID = uuid.New()
	return nil
}
