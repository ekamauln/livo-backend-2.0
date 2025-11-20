package models

import (
	"time"

	"gorm.io/gorm"
)

type Expedition struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Code      string         `gorm:"size:4;not null;unique;check:code_len_check,length(code) >= 2" json:"code"`
	Name      string         `gorm:"not null" json:"name"`
	Slug      string         `gorm:"not null" json:"slug"`
	Color     string         `json:"color"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type ExpeditionResponse struct {
	ID      uint      `json:"id"`
	Code    string    `json:"code"`
	Name    string    `json:"name"`
	Slug    string    `json:"slug"`
	Color   string    `json:"color"`
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
}

// ToExpeditionResponse converts Expedition model to ExpeditionResponse
func (e *Expedition) ToExpeditionResponse() ExpeditionResponse {
	return ExpeditionResponse{
		ID:      e.ID,
		Code:    e.Code,
		Name:    e.Name,
		Slug:    e.Slug,
		Color:   e.Color,
		Created: e.CreatedAt,
		Updated: e.UpdatedAt,
	}
}
