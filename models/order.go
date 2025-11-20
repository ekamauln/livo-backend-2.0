package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	OrderGineeID    string         `gorm:"unique;not null" json:"order_ginee_id"`
	Status          string         `gorm:"not null" json:"status"`
	Type            string         `gorm:"not null" json:"type"`
	Channel         string         `gorm:"not null" json:"channel"`
	Store           string         `gorm:"not null" json:"store"`
	Buyer           string         `gorm:"not null" json:"buyer"`
	Address         string         `gorm:"not null" json:"address"`
	Courier         string         `gorm:"not null" json:"courier"`
	Tracking        string         `gorm:"unique;not null" json:"tracking"`
	ImporterID      *uint          `gorm:"default:null" json:"importer_id"`
	PickerID        *uint          `gorm:"default:null" json:"picker_id"`
	UpdaterID       *uint          `gorm:"default:null" json:"updater_id"`
	CancelerID      *uint          `gorm:"default:null" json:"canceler_id"`
	Complained      bool           `gorm:"default:false" json:"complained"`
	PickedAt        *time.Time     `gorm:"default:null" json:"picked_at"`
	ProcessingLimit time.Time      `gorm:"not null" json:"processing_limit"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	CancelAt        *time.Time     `gorm:"default:null" json:"cancel_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	OrderDetails []OrderDetail `gorm:"foreignKey:OrderID" json:"order_details"`
	Picker       *User         `gorm:"foreignKey:PickerID" json:"picker,omitempty"`
	Importer     *User         `gorm:"foreignKey:ImporterID" json:"importer,omitempty"`
	Updater      *User         `gorm:"foreignKey:UpdaterID" json:"updater,omitempty"`
	Canceler     *User         `gorm:"foreignKey:CancelerID" json:"canceler,omitempty"`
}

type OrderDetail struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	OrderID     uint           `json:"order_id"`
	Sku         string         `json:"sku" gorm:"index"`
	ProductName string         `json:"product_name"`
	Variant     string         `json:"variant"`
	Quantity    int            `json:"quantity"`
	Product     *Product       `json:"product,omitempty" gorm:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// OrderResponse represents order data for API responses
type OrderResponse struct {
	ID           uint                  `json:"id"`
	OrderGineeID string                `json:"order_ginee_id"`
	Status       string                `json:"status"`
	Channel      string                `json:"channel"`
	Store        string                `json:"store"`
	Buyer        string                `json:"buyer"`
	Courier      string                `json:"courier"`
	Tracking     string                `json:"tracking"`
	Complained   bool                  `json:"complained"`
	ImportedBy   string                `json:"imported_by"`
	UpdatedBy    string                `json:"updated_by"`
	PickedBy     string                `json:"picked_by"`
	CanceledBy   string                `json:"canceled_by"`
	PickedAt     string                `json:"picked_at"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	CancelAt     string                `json:"cancel_at"`
	OrderDetails []OrderDetailResponse `json:"order_details"`
}

type OrderDetailResponse struct {
	ID          uint             `json:"id"`
	Sku         string           `json:"sku"`
	ProductName string           `json:"product_name"`
	Variant     string           `json:"variant"`
	Quantity    int              `json:"quantity"`
	Product     *ProductResponse `json:"product,omitempty"`
}

// ToOrderResponse converts Order model to OrderResponse
func (o *Order) ToOrderResponse() OrderResponse {
	details := make([]OrderDetailResponse, len(o.OrderDetails))
	for i, detail := range o.OrderDetails {
		detailResp := OrderDetailResponse{
			ID:          detail.ID,
			Sku:         detail.Sku,
			ProductName: detail.ProductName,
			Variant:     detail.Variant,
			Quantity:    detail.Quantity,
		}

		// Include product data if exists
		if detail.Product != nil {
			detailResp.Product = &ProductResponse{
				ID:    detail.Product.ID,
				Sku:   detail.Product.Sku,
				Name:  detail.Product.Name,
				Image: detail.Product.Image,
			}
		}

		details[i] = detailResp
	}

	// Handle picked_at field
	var pickedAtStr string
	if o.PickedAt != nil {
		pickedAtStr = o.PickedAt.Format("2006-01-02 15:04:05")
	} else {
		pickedAtStr = "Not picked yet"
	}

	// Handle canceled_at field
	var canceledAtStr string
	if o.CancelAt != nil {
		canceledAtStr = o.CancelAt.Format("2006-01-02 15:04:05")
	} else {
		canceledAtStr = "Not canceled"
	}

	// Handle canceled_by field
	var canceledByStr string
	if o.Canceler != nil {
		canceledByStr = o.Canceler.Name + " (" + o.Canceler.Username + ")"
	} else {
		canceledByStr = "Not canceled"
	}

	// Handle picked_by field
	var pickedByStr string
	if o.Picker != nil {
		pickedByStr = o.Picker.Name + " (" + o.Picker.Username + ")"
	} else {
		pickedByStr = "Not picked yet"
	}

	// Handle imported_by field
	var importedByStr string
	if o.Importer != nil {
		importedByStr = o.Importer.Name + " (" + o.Importer.Username + ")"
	} else {
		importedByStr = "Not imported yet"
	}

	// Handle updated_by field
	var updatedByStr string
	if o.Updater != nil {
		updatedByStr = o.Updater.Name + " (" + o.Updater.Username + ")"
	} else {
		updatedByStr = "Not updated yet"
	}

	return OrderResponse{
		ID:           o.ID,
		OrderGineeID: o.OrderGineeID,
		Status:       o.Status,
		Channel:      o.Channel,
		Store:        o.Store,
		Buyer:        o.Buyer,
		Courier:      o.Courier,
		Tracking:     o.Tracking,
		Complained:   o.Complained,
		ImportedBy:   importedByStr,
		UpdatedBy:    updatedByStr,
		PickedBy:     pickedByStr,
		CanceledBy:   canceledByStr,
		PickedAt:     pickedAtStr,
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
		CancelAt:     canceledAtStr,
		OrderDetails: details,
	}
}
