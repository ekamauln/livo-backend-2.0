package controllers

import (
	"net/http"

	"livo-backend-2.0/config"
	"livo-backend-2.0/models"
	"livo-backend-2.0/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	DB     *gorm.DB
	Config *config.Config
}

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"budi"`
	Email    string `json:"email" binding:"required,email" example:"budi@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
	Name     string `json:"name" binding:"required" example:"Budiawan Bengi"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"budi"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	AccessToken  string              `json:"access_token"`
	RefreshToken string              `json:"refresh_token"`
	User         models.UserResponse `json:"user"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// NewAuthController creates a new auth controller
func NewAuthController(db *gorm.DB, config *config.Config) *AuthController {
	return &AuthController{
		DB:     db,
		Config: config,
	}
}

// Register godoc
// @Summary Pendaftaran pengguna baru
// @Description Mendaftarkan akun pengguna baru
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 201 {object} utils.Response{data=models.UserResponse}
// @Failure 400 {object} utils.Response
// @Failure 409 {object} utils.Response
// @Router /api/auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := ac.DB.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Username atau email sudah digunakan", "username atau email tersebut sudah digunakan")
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengenkripsi password", err.Error())
		return
	}

	// Create user
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
		IsActive: true,
	}

	if err := ac.DB.Create(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat pengguna", err.Error())
		return
	}

	// Assign guest role by default
	var guestRole models.Role
	if err := ac.DB.Where("name = ?", "guest").First(&guestRole).Error; err == nil {
		userRole := models.UserRole{
			UserID:     user.ID,
			RoleID:     guestRole.ID,
			AssignedBy: 1,
		}
		ac.DB.Create(&userRole)
	}

	// Load user with roles
	ac.DB.Preload("UserRoles.Role").First(&user, user.ID)

	utils.SuccessResponse(c, http.StatusCreated, "Pengguna berhasil didaftarkan", user.ToUserResponse())
}

// Login godoc
// @Summary Login user
// @Description Mengautentikasi pengguna dan mengembalikan token JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} utils.Response{data=LoginResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Find user
	var user models.User
	if err := ac.DB.Preload("UserRoles.Role").Where("username = ?", req.Username).First(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Username salah", "user tidak ditemukan")
		return
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Password salah", "password salah")
		return
	}

	// Check if user is active
	if !user.IsActive {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Akun tidak aktif", "akun pengguna dinonaktifkan")
		return
	}

	// Extract roles
	roles := make([]string, len(user.UserRoles))
	for i, userRole := range user.UserRoles {
		roles[i] = userRole.Role.Role
	}

	// Generate tokens
	accessToken, refreshToken, err := utils.GenerateTokens(
		user.ID,
		user.Username,
		roles,
		ac.Config.JWTSecret,
		ac.Config.JWTExpireHours,
		ac.Config.RefreshTokenExpireDays,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghasilkan token", err.Error())
		return
	}

	// Save refresh token
	user.RefreshToken = refreshToken
	ac.DB.Save(&user)

	response := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToUserResponse(),
	}

	utils.SuccessResponse(c, http.StatusOK, "Login sukses", response)
}

// RefreshToken godoc
// @Summary Memperbarui access token
// @Description Memperbarui access token baru menggunakan token refresh
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} utils.Response{data=LoginResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/auth/refresh [post]
func (ac *AuthController) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate refresh token
	claims, err := utils.ValidateRefreshToken(req.RefreshToken, ac.Config.JWTSecret)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Refresh token salah", err.Error())
		return
	}

	// Find user
	var user models.User
	if err := ac.DB.Preload("UserRoles.Role").Preload("UserRoles.Assigner").Where("id = ? AND refresh_token = ?", claims.UserID, req.RefreshToken).First(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Refresh token salah", "refresh token tidak ditemukan untuk pengguna ini")
		return
	}

	// Extract roles
	roles := make([]string, len(user.UserRoles))
	for i, userRole := range user.UserRoles {
		roles[i] = userRole.Role.Role
	}

	// Generate new tokens
	accessToken, refreshToken, err := utils.GenerateTokens(
		user.ID,
		user.Username,
		roles,
		ac.Config.JWTSecret,
		ac.Config.JWTExpireHours,
		ac.Config.RefreshTokenExpireDays,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghasilkan access token", err.Error())
		return
	}

	// Update refresh token
	user.RefreshToken = refreshToken
	ac.DB.Save(&user)

	response := LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToUserResponse(),
	}

	utils.SuccessResponse(c, http.StatusOK, "Token berhasil diperbarui", response)
}

// Logout godoc
// @Summary Logout user
// @Description Logout user dengan menonaktifkan refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Clear refresh token
	if err := ac.DB.Model(&models.User{}).Where("id = ?", userID).Update("refresh_token", "").Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal logout", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Logout sukses", nil)
}
