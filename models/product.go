package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Sku       string         `gorm:"unique;not null" json:"sku"`
	Name      string         `gorm:"not null" json:"name"`
	Image     string         `json:"image"`
	Variant   string         `json:"variant"`
	Location  string         `json:"location"`
	Barcode   string         `json:"barcode"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type ProductResponse struct {
	ID       uint      `json:"id"`
	Sku      string    `json:"sku"`
	Name     string    `json:"name"`
	Image    string    `json:"image"`
	Variant  string    `json:"variant"`
	Location string    `json:"location"`
	Barcode  string    `json:"barcode"`
	Created  time.Time `json:"created_at"`
	Updated  time.Time `json:"updated_at"`
}

// ToProductResponse converts Product model to ProductResponse
func (p *Product) ToProductResponse() ProductResponse {
	return ProductResponse{
		ID:       p.ID,
		Sku:      p.Sku,
		Name:     p.Name,
		Image:    p.Image,
		Variant:  p.Variant,
		Location: p.Location,
		Barcode:  p.Barcode,
		Created:  p.CreatedAt,
		Updated:  p.UpdatedAt,
	}
}
