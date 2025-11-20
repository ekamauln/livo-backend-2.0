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

type StoreController struct {
	DB *gorm.DB
}

// NewStoreController creates a new store controller
func NewStoreController(db *gorm.DB) *StoreController {
	return &StoreController{DB: db}
}

// GetStores godoc
// @Summary Get all stores
// @Description Mengambil data semua store dengan pagination dan pencarian.
// @Tags stores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by Code or Name (partial match)"
// @Success 200 {object} utils.Response{data=StoresListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/stores [get]
func (sc *StoreController) GetStores(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Parse search parameter
	search := c.Query("search")

	var stores []models.Store
	var total int64

	// Build query with optional search
	query := sc.DB.Model(&models.Store{})

	if search != "" {
		// Search by Code or Name with partial match
		query = query.Where("code ILIKE ? OR name ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count with search filter
	if err := query.Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghitung jumlah stores", err.Error())
		return
	}

	// Get stores with pagination, search filter, and order by ID ascending
	if err := query.Order("id ASC").Limit(limit).Offset(offset).Find(&stores).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data stores", err.Error())
		return
	}

	// Convert to response format
	storeResponses := make([]models.StoreResponse, len(stores))
	for i, store := range stores {
		storeResponses[i] = store.ToStoreResponse()
	}

	response := StoresListResponse{
		Stores: storeResponses,
		Pagination: utils.PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}

	// Build success message
	message := "Stores berhasil diambil"
	if search != "" {
		message += " (difilter berdasarkan kode atau nama: " + search + ")"
	}

	utils.SuccessResponse(c, http.StatusOK, message, response)
}

// GetStore godoc
// @Summary Get store by ID
// @Description Mengambil data store spesifik berdasarkan ID.
// @Tags stores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Store ID"
// @Success 200 {object} utils.Response{data=models.StoreResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/stores/{id} [get]
func (sc *StoreController) GetStore(c *gin.Context) {
	storeID := c.Param("id")

	var store models.Store
	if err := sc.DB.First(&store, storeID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Store tidak ditemukan", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Store berhasil diambil", store.ToStoreResponse())
}

// UpdateStore godoc
// @Summary Update store
// @Description Memperbarui data store.
// @Tags stores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Store ID"
// @Param store body UpdateStoreRequest true "Update Store Request"
// @Success 200 {object} utils.Response{data=models.StoreResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/stores/{id} [put]
func (sc *StoreController) UpdateStore(c *gin.Context) {
	storeID := c.Param("id")

	var req UpdateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	var store models.Store
	if err := sc.DB.First(&store, storeID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Store tidak ditemukan", err.Error())
		return
	}

	// Check for duplicate store code (excluding current store)
	var existingStore models.Store
	if err := sc.DB.Where("code = ? AND id <> ?", req.Code, store.ID).First(&existingStore).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Kode store sudah ada", "Store dengan kode ini sudah ada")
		return
	}

	// Update store fields
	store.Code = req.Code
	store.Name = req.Name

	if err := sc.DB.Save(&store).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memperbarui store", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Store berhasil diperbarui", store.ToStoreResponse())
}

// RemoveStore godoc
// @Summary Remove store
// @Description Menghapus data store.
// @Tags stores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Store ID"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /api/stores/{id} [delete]
func (sc *StoreController) RemoveStore(c *gin.Context) {
	storeID := c.Param("id")

	var store models.Store
	if err := sc.DB.First(&store, storeID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Store tidak ditemukan", err.Error())
		return
	}

	if err := sc.DB.Delete(&store).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal menghapus store", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Store berhasil dihapus", nil)
}

// CreateStore godoc
// @Summary Create new store
// @Description Membuat store baru.
// @Tags stores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param store body CreateStoreRequest true "Create Store Request"
// @Success 201 {object} utils.Response{data=models.StoreResponse}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/stores [post]
func (sc *StoreController) CreateStore(c *gin.Context) {
	var req CreateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// convert code to uppercase and trim spaces
	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))

	store := models.Store{
		Code: req.Code,
		Name: req.Name,
	}
	// Check for duplicate store code
	var existingStore models.Store
	if err := sc.DB.Where("code = ?", req.Code).First(&existingStore).Error; err == nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Kode store sudah ada", "Store dengan kode ini sudah ada")
		return
	}

	// Create new store and return response
	if err := sc.DB.Create(&store).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat store", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Store berhasil dibuat", store.ToStoreResponse())
}

// Request/Response structs
type StoresListResponse struct {
	Stores     []models.StoreResponse   `json:"stores"`
	Pagination utils.PaginationResponse `json:"pagination"`
}

type UpdateStoreRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type CreateStoreRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}
