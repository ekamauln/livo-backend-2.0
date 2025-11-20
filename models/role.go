package models

import (
	"time"

	"gorm.io/gorm"
)

// Role represents system roles
type Role struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Role        string         `gorm:"unique;not null" json:"role"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// GetRoleHierarchy returns role hierarchy levels
func GetRoleHierarchy() map[string]int {
	return map[string]int{
		"superadmin":  9,
		"coordinator": 4,
		"admin":       3,
		"admin-retur": 3,
		"finance":     3,
		"warehouse":   3,
		"picker":      2,
		"outbound":    2,
		"qc-ribbon":   2,
		"qc-online":   2,
		"mb-ribbon":   2,
		"mb-online":   2,
		"packing":     2,
		"guest":       1,
	}
}

// CanManageRole checks if a role can manage another role
func (r *Role) CanManageRole(targetRole string) bool {
	hierarchy := GetRoleHierarchy()
	currentLevel, exists := hierarchy[r.Role]
	if !exists {
		return false
	}

	targetLevel, exists := hierarchy[targetRole]
	if !exists {
		return false
	}

	return currentLevel > targetLevel
}

type RoleListResponse struct {
	ID          uint      `json:"id"`
	Role        string    `json:"role"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToRoleListResponse converts Role model to RoleListResponse
func (r *Role) ToRoleListResponse() RoleListResponse {
	return RoleListResponse{
		ID:          r.ID,
		Role:        r.Role,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
