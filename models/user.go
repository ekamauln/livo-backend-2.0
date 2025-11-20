package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"unique;not null" json:"username"`
	Email        string         `gorm:"unique;not null" json:"email"`
	Password     string         `gorm:"not null" json:"-"`
	Name         string         `gorm:"not null" json:"name"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	RefreshToken string         `json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	UserRoles []UserRole `gorm:"foreignKey:UserID" json:"user_roles"`
}

// UserResponse represents user data for API responses
type UserResponse struct {
	ID        uint           `json:"id"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	Name      string         `json:"name"`
	IsActive  bool           `json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Roles     []RoleResponse `json:"roles"`
}

// RoleResponse represents role data for API responses
type RoleResponse struct {
	ID          uint   `json:"id"`
	Role        string `json:"role"`
	Description string `json:"description"`
	AssignedBy  string `json:"assigned_by"`
	AssignedAt  string `json:"assigned_at"`
}

// ToUserResponse converts User model to UserResponse
func (u *User) ToUserResponse() UserResponse {
	roles := make([]RoleResponse, len(u.UserRoles))
	for i, ur := range u.UserRoles {
		roles[i] = RoleResponse{
			ID:          ur.Role.ID,
			Role:        ur.Role.Role,
			Description: ur.Role.Description,
			AssignedBy:  ur.Assigner.Username,
			AssignedAt:  ur.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Name:      u.Name,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Roles:     roles,
	}
}

// HasRole checks if user has a specific role
func (u *User) HasRole(roleName string) bool {
	for _, userRole := range u.UserRoles {
		if userRole.Role.Role == roleName {
			return true
		}
	}
	return false
}

// GetHighestRoleLevel returns the highest role level of the user
func (u *User) GetHighestRoleLevel() int {
	hierarchy := GetRoleHierarchy()
	maxLevel := 0

	for _, userRole := range u.UserRoles {
		if level, exists := hierarchy[userRole.Role.Role]; exists {
			if level > maxLevel {
				maxLevel = level
			}
		}
	}

	return maxLevel
}

// CanManageUser checks if current user can manage another user
func (u *User) CanManageUser(targetUser *User) bool {
	currentLevel := u.GetHighestRoleLevel()
	targetLevel := targetUser.GetHighestRoleLevel()

	return currentLevel > targetLevel
}
