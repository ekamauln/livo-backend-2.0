package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"livo-backend-2.0/models"
	"livo-backend-2.0/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserManagerController struct {
	DB *gorm.DB
}

// NewUserManagerController creates a new user manager controller
func NewUserManagerController(db *gorm.DB) *UserManagerController {
	return &UserManagerController{DB: db}
}

// GetUsers godoc
// @Summary Get all users
// @Description Mengambil daftar semua user dengan kemampuan pencarian opsional berdasarkan username atau nama.
// @Tags user-manager
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Pencarian berdasarkan username atau nama"
// @Success 200 {object} utils.Response{data=UsersListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/user-manager/users [get]
func (ac *UserManagerController) GetUsers(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Parse search parameter
	search := strings.TrimSpace(c.Query("search"))

	var users []models.User
	var total int64

	// Build base query
	query := ac.DB.Model(&models.User{})

	// Add search conditions if search parameter is provided
	if search != "" {
		searchCondition := "username ILIKE ? OR name ILIKE ?"
		searchPattern := "%" + search + "%"
		query = query.Where(searchCondition, searchPattern, searchPattern)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghitung jumlah user", err.Error())
		return
	}

	// Get users with pagination and order by ID ascending
	if err := query.Order("id ASC").Preload("UserRoles.Role").Preload("UserRoles.Assigner").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data user", err.Error())
		return
	}

	// Convert to response format
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToUserResponse()
	}

	response := UsersListResponse{
		Users: userResponses,
		Pagination: utils.PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil data semua user", response)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Mengambil informasi user spesifik berdasarkan ID user.
// @Tags user-manager
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} utils.Response{data=models.UserResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/user-manager/users/{id} [get]
func (ac *UserManagerController) GetUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := ac.DB.Preload("UserRoles.Role").First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User tidak ditemukan", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil data user", user.ToUserResponse())
}

// UpdateUserStatus godoc
// @Summary Update user status (active/inactive)
// @Description Mengaktifkan atau menonaktifkan status user. (hanya coordinator yang dapat mengakses)
// @Tags user-manager
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body UpdateUserStatusRequest true "Update status request"
// @Success 200 {object} utils.Response{data=models.UserResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/user-manager/users/{id}/status [put]
func (ac *UserManagerController) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("id")

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	var user models.User
	if err := ac.DB.First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User tidak ditemukan", err.Error())
		return
	}

	user.IsActive = req.IsActive
	if err := ac.DB.Save(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui status user", err.Error())
		return
	}

	// Load user with roles
	ac.DB.Preload("UserRoles.Role").First(&user, user.ID)

	utils.SuccessResponse(c, http.StatusOK, "Berhasil memperbarui status user", user.ToUserResponse())
}

// AssignRole godoc
// @Summary Assign role to user
// @Description Menugaskan peran role ke user (hanya coordinator yang dapat mengakses)
// @Tags user-manager
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body AssignRoleRequest true "Assign role request"
// @Success 200 {object} utils.Response{data=models.UserResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/user-manager/users/{id}/roles [post]
func (ac *UserManagerController) AssignRole(c *gin.Context) {
	userID := c.Param("id")

	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Find target user
	var user models.User
	if err := ac.DB.Preload("UserRoles.Role").First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User tidak ditemukan", err.Error())
		return
	}

	// Find role
	var role models.Role
	if err := ac.DB.Where("name = ?", req.RoleName).First(&role).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Role tidak ditemukan", err.Error())
		return
	}

	// Check if user already has this role
	for _, userRole := range user.UserRoles {
		if userRole.RoleID == role.ID {
			utils.ErrorResponse(c, http.StatusConflict, "User sudah memiliki role ini", "role sudah ditugaskan sebelumnya")
			return
		}
	}

	// Check permission hierarchy (get current user's roles)
	currentUserRoles, _ := c.Get("roles")
	currentRoles := currentUserRoles.([]string)

	// Get current user's highest role level
	hierarchy := models.GetRoleHierarchy()
	currentMaxLevel := 0
	for _, roleName := range currentRoles {
		if level, exists := hierarchy[roleName]; exists && level > currentMaxLevel {
			currentMaxLevel = level
		}
	}

	// Check if current user can assign this role
	targetRoleLevel, exists := hierarchy[req.RoleName]
	if !exists || currentMaxLevel < targetRoleLevel {
		utils.ErrorResponse(c, http.StatusForbidden, "Tidak memiliki izin untuk menugaskan role ini", "izin ditolak")
		return
	}

	// Get current user ID from context
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User tidak terautentikasi", "user_id tidak ditemukan dalam konteks")
		return
	}

	// Assign role
	userRole := models.UserRole{
		UserID:     user.ID,
		RoleID:     role.ID,
		AssignedBy: currentUserID.(uint),
	}

	if err := ac.DB.Create(&userRole).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menugaskan role", err.Error())
		return
	}

	// Reload user with updated roles
	ac.DB.Preload("UserRoles.Role").First(&user, user.ID)

	utils.SuccessResponse(c, http.StatusOK, "Berhasil menugaskan role ke user", user.ToUserResponse())
}

// RemoveRole godoc
// @Summary Remove role from user
// @Description Menghapus peran role dari user (hanya coordinator yang dapat mengakses)
// @Tags user-manager
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body RemoveRoleRequest true "Remove role request"
// @Success 200 {object} utils.Response{data=models.UserResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/user-manager/users/{id}/roles [delete]
func (ac *UserManagerController) RemoveRole(c *gin.Context) {
	userID := c.Param("id")

	var req RemoveRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Find role
	var role models.Role
	if err := ac.DB.Where("name = ?", req.RoleName).First(&role).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Role tidak ditemukan", err.Error())
		return
	}

	// Check permission hierarchy
	currentUserRoles, _ := c.Get("roles")
	currentRoles := currentUserRoles.([]string)

	hierarchy := models.GetRoleHierarchy()
	currentMaxLevel := 0
	for _, roleName := range currentRoles {
		if level, exists := hierarchy[roleName]; exists && level > currentMaxLevel {
			currentMaxLevel = level
		}
	}

	targetRoleLevel, exists := hierarchy[req.RoleName]
	if !exists || currentMaxLevel < targetRoleLevel {
		utils.ErrorResponse(c, http.StatusForbidden, "Tidak memiliki izin untuk menghapus role ini", "izin ditolak")
		return
	}

	// Remove role
	if err := ac.DB.Where("user_id = ? AND role_id = ?", userID, role.ID).Delete(&models.UserRole{}).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus role", err.Error())
		return
	}

	// Reload user with updated roles
	var user models.User
	ac.DB.Preload("UserRoles.Role").Preload("UserRoles.Assigner").First(&user, userID)

	utils.SuccessResponse(c, http.StatusOK, "Berhasil menghapus role dari user", user.ToUserResponse())
}

// CreateUser godoc
// @Summary Create new user
// @Description Membuat akun baru untuk user (hanya coordinator yang dapat mengakses)
// @Tags user-manager
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateUserRequest true "Create user request"
// @Success 201 {object} utils.Response{data=models.UserResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Router /api/user-manager/users [post]
func (ac *UserManagerController) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := ac.DB.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "User sudah ada", "username atau email sudah digunakan")
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengenkripsi password", err.Error())
		return
	}

	// Get current user ID for audit trail
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User tidak terautentikasi", "user_id tidak ditemukan dalam konteks")
		return
	}

	// Create user
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
		IsActive: req.IsActive,
	}

	if err := ac.DB.Create(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat user", err.Error())
		return
	}

	// Assign initial role if specified
	if req.InitialRole != "" {
		// Check permission hierarchy for role assignment
		currentUserRoles, _ := c.Get("roles")
		currentRoles := currentUserRoles.([]string)

		hierarchy := models.GetRoleHierarchy()
		currentMaxLevel := 0
		for _, roleName := range currentRoles {
			if level, exists := hierarchy[roleName]; exists && level > currentMaxLevel {
				currentMaxLevel = level
			}
		}

		// Check if current user can assign this role
		targetRoleLevel, exists := hierarchy[req.InitialRole]
		if !exists {
			utils.ErrorResponse(c, http.StatusBadRequest, "Role tidak valid", "role tidak ditemukan")
			return
		}

		if currentMaxLevel < targetRoleLevel {
			utils.ErrorResponse(c, http.StatusForbidden, "Tidak memiliki izin untuk menetapkan role ini", "izin ditolak")
			return
		}

		// Find and assign the role
		var role models.Role
		if err := ac.DB.Where("name = ?", req.InitialRole).First(&role).Error; err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Role tidak ditemukan", err.Error())
			return
		}

		userRole := models.UserRole{
			UserID:     user.ID,
			RoleID:     role.ID,
			AssignedBy: currentUserID.(uint),
		}

		if err := ac.DB.Create(&userRole).Error; err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menetapkan role", err.Error())
			return
		}
	} else {
		// Assign guest role by default
		var guestRole models.Role
		if err := ac.DB.Where("name = ?", "guest").First(&guestRole).Error; err == nil {
			userRole := models.UserRole{
				UserID:     user.ID,
				RoleID:     guestRole.ID,
				AssignedBy: currentUserID.(uint),
			}
			ac.DB.Create(&userRole)
		}
	}

	// Load user with roles
	ac.DB.Preload("UserRoles.Role").Preload("UserRoles.Assigner").First(&user, user.ID)

	utils.SuccessResponse(c, http.StatusCreated, "User berhasil dibuat", user.ToUserResponse())
}

// DeleteUser godoc
// @Summary Remove user account
// @Description Menghapus akun user. (hanya coordinator yang dapat mengakses)
// @Tags user-manager
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/user-manager/users/{id} [delete]
func (ac *UserManagerController) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// Find user to be deleted
	var user models.User
	if err := ac.DB.Preload("UserRoles.Role").First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User tidak ditemukan", err.Error())
		return
	}

	// Prevent deletion of current user
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User tidak terautentikasi", "user_id tidak ditemukan dalam konteks")
		return
	}

	if user.ID == currentUserID.(uint) {
		utils.ErrorResponse(c, http.StatusForbidden, "Tidak dapat menghapus akun sendiri", "penghapusan diri tidak diizinkan")
		return
	}

	// Check permission hierarchy - can only delete users with lower roles
	currentUserRoles, _ := c.Get("roles")
	currentRoles := currentUserRoles.([]string)

	hierarchy := models.GetRoleHierarchy()
	currentMaxLevel := 0
	for _, roleName := range currentRoles {
		if level, exists := hierarchy[roleName]; exists && level > currentMaxLevel {
			currentMaxLevel = level
		}
	}

	// Get target user's highest role level
	targetMaxLevel := 0
	for _, userRole := range user.UserRoles {
		if level, exists := hierarchy[userRole.Role.Role]; exists && level > targetMaxLevel {
			targetMaxLevel = level
		}
	}

	// Check if current user has permission to delete target user
	if currentMaxLevel <= targetMaxLevel {
		utils.ErrorResponse(c, http.StatusForbidden, "Tidak memiliki izin untuk menghapus user ini", "izin ditolak")
		return
	}

	// Start transaction to ensure data consistency
	tx := ac.DB.Begin()

	// Delete all user roles first (due to foreign key constraints)
	if err := tx.Where("user_id = ?", user.ID).Delete(&models.UserRole{}).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus peran user", err.Error())
		return
	}

	// Delete the user (soft delete)
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus user", err.Error())
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal melakukan commit transaksi", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User berhasil dihapus", nil)
}

// UpdateUserPassword godoc
// @Summary Update user password
// @Description Memperbarui kata sandi user (hanya coordinator yang dapat mengakses)
// @Tags user-manager
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body UpdateUserPasswordRequest true "Update password request"
// @Success 200 {object} utils.Response{data=models.UserResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/user-manager/users/{id}/password [put]
func (ac *UserManagerController) UpdateUserPassword(c *gin.Context) {
	userID := c.Param("id")

	var req UpdateUserPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Find user to be updated
	var user models.User
	if err := ac.DB.Preload("UserRoles.Role").First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User tidak ditemukan", err.Error())
		return
	}

	// Check permission hierarchy - can only update users with lower or equal roles
	currentUserRoles, _ := c.Get("roles")
	currentRoles := currentUserRoles.([]string)

	hierarchy := models.GetRoleHierarchy()
	currentMaxLevel := 0
	for _, roleName := range currentRoles {
		if level, exists := hierarchy[roleName]; exists && level > currentMaxLevel {
			currentMaxLevel = level
		}
	}

	// Get target user's highest role level
	targetMaxLevel := 0
	for _, userRole := range user.UserRoles {
		if level, exists := hierarchy[userRole.Role.Role]; exists && level > targetMaxLevel {
			targetMaxLevel = level
		}
	}

	// Check if current user has permission to update target user
	if currentMaxLevel < targetMaxLevel {
		utils.ErrorResponse(c, http.StatusForbidden, "Tidak memiliki izin untuk memperbarui user ini", "izin ditolak")
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengenkripsi kata sandi", err.Error())
		return
	}

	// Update password
	user.Password = hashedPassword
	if err := ac.DB.Save(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui kata sandi", err.Error())
		return
	}

	// Clear refresh token to force re-login
	user.RefreshToken = ""
	ac.DB.Save(&user)

	// Load user with roles for response
	ac.DB.Preload("UserRoles.Role").Preload("UserRoles.Assigner").First(&user, user.ID)

	utils.SuccessResponse(c, http.StatusOK, "Kata sandi berhasil diperbarui", user.ToUserResponse())
}

// UpdateUserProfile godoc
// @Summary Update user profile
// @Description Memperbarui nama lengkap dan email user (hanya coordinator yang dapat mengakses)
// @Tags user-manager
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body UpdateUserProfileRequest true "Update profile request"
// @Success 200 {object} utils.Response{data=models.UserResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Router /api/user-manager/users/{id}/profile [put]
func (ac *UserManagerController) UpdateUserProfile(c *gin.Context) {
	userID := c.Param("id")

	var req UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Find user to be updated
	var user models.User
	if err := ac.DB.Preload("UserRoles.Role").First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User tidak ditemukan", err.Error())
		return
	}

	// Check if email already exists (if email is being changed)
	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		if err := ac.DB.Where("email = ? AND id != ?", req.Email, user.ID).First(&existingUser).Error; err == nil {
			utils.ErrorResponse(c, http.StatusConflict, "Email sudah digunakan", "email sudah digunakan oleh user lain")
			return
		}
	}

	// Check permission hierarchy - can only update users with lower or equal roles
	currentUserRoles, _ := c.Get("roles")
	currentRoles := currentUserRoles.([]string)

	hierarchy := models.GetRoleHierarchy()
	currentMaxLevel := 0
	for _, roleName := range currentRoles {
		if level, exists := hierarchy[roleName]; exists && level > currentMaxLevel {
			currentMaxLevel = level
		}
	}

	// Get target user's highest role level
	targetMaxLevel := 0
	for _, userRole := range user.UserRoles {
		if level, exists := hierarchy[userRole.Role.Role]; exists && level > targetMaxLevel {
			targetMaxLevel = level
		}
	}

	// Check if current user has permission to update target user
	if currentMaxLevel < targetMaxLevel {
		utils.ErrorResponse(c, http.StatusForbidden, "Tidak memiliki izin untuk memperbarui user ini", "izin ditolak")
		return
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	// Save changes
	if err := ac.DB.Save(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui profil user", err.Error())
		return
	}

	// Load user with roles for response
	ac.DB.Preload("UserRoles.Role").Preload("UserRoles.Assigner").First(&user, user.ID)

	utils.SuccessResponse(c, http.StatusOK, "Profil user berhasil diperbarui", user.ToUserResponse())
}

// GetRoles godoc
// @Summary Get all roles
// @Description Mengambil data semua role yang tersedia. (hanya coordinator yang dapat mengakses)
// @Tags user-manager
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} utils.Response{data=[]models.Role}
// @Failure 401 {object} utils.Response
// @Router /api/user-manager/roles [get]
func (ac *UserManagerController) GetRoles(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var roles []models.Role
	var total int64

	// Get total count
	ac.DB.Model(&models.Role{}).Count(&total)

	if err := ac.DB.Limit(limit).Offset(offset).Find(&roles).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data role", err.Error())
		return
	}

	// Convert to response format
	roleResponses := make([]models.RoleListResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = role.ToRoleListResponse()
	}

	response := RoleListResponse{
		Roles: roleResponses,
		Pagination: utils.PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Berhasil mengambil data role", response)
}

// Request/Response structs
type UsersListResponse struct {
	Users      []models.UserResponse    `json:"users"`
	Pagination utils.PaginationResponse `json:"pagination"`
}

type RoleListResponse struct {
	Roles      []models.RoleListResponse `json:"roles"`
	Pagination utils.PaginationResponse  `json:"pagination"`
}

type CreateUserRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50" example:"budi"`
	Email       string `json:"email" binding:"required,email" example:"budi@example.com"`
	Password    string `json:"password" binding:"required,min=6" example:"password123"`
	Name        string `json:"name" binding:"required" example:"Budi Santoso"`
	IsActive    bool   `json:"is_active" example:"true"`
	InitialRole string `json:"initial_role,omitempty" example:"picker"`
}

type UpdateUserStatusRequest struct {
	IsActive bool `json:"is_active" example:"true"`
}

type UpdateUserPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
}

type UpdateUserProfileRequest struct {
	Name  string `json:"name,omitempty" binding:"omitempty,min=1" example:"Budi Santoso Updated"`
	Email string `json:"email,omitempty" binding:"omitempty,email" example:"newemail@example.com"`
}

type AssignRoleRequest struct {
	RoleName string `json:"role_name" binding:"required" example:"manager"`
}

type RemoveRoleRequest struct {
	RoleName string `json:"role_name" binding:"required" example:"manager"`
}
