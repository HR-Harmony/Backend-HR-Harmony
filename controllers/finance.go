package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func CreateFinanceByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind the finance data from the request body
		var finance models.Finance
		if err := c.Bind(&finance); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate finance data
		if finance.AccountTitle == "" || finance.InitialBalance == 0 || finance.AccountNumber == "" || finance.BranchCode == "" || finance.BankBranch == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid finance data. All fields are required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Create the finance in the database
		db.Create(&finance)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Finance data added successfully",
			"data":    finance,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllFinanceByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Fetch searching query parameter
		searching := c.QueryParam("searching")

		// Pagination parameters
		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10 // Default per page
		}

		var finances []models.Finance
		query := db.Model(&finances)
		if searching != "" {
			query = query.Where("LOWER(account_title) ILIKE ? OR cast(initial_balance as text) LIKE ? OR account_number LIKE ? OR branch_code LIKE ? OR LOWER(bank_branch) ILIKE ?", "%"+strings.ToLower(searching)+"%", "%"+searching+"%", "%"+searching+"%", "%"+searching+"%", "%"+strings.ToLower(searching)+"%")
		}

		// Count total records for pagination
		var totalCount int64
		query.Count(&totalCount)

		// Calculate offset and limit for pagination
		offset := (page - 1) * perPage

		// Fetch data with pagination
		query.Offset(offset).Limit(perPage).Find(&finances)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Finance data retrieved successfully",
			"data":       finances,
			"pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetFinanceByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Extract finance ID from the request
		financeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid finance ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch finance data from the database
		var finance models.Finance
		result = db.First(&finance, uint(financeID))
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Finance data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Finance data retrieved successfully",
			"data":    finance,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateFinanceByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Extract finance ID from the request
		financeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid finance ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch finance data from the database
		var finance models.Finance
		result = db.First(&finance, uint(financeID))
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Finance data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated finance data from the request body
		var updatedFinance models.Finance
		if err := c.Bind(&updatedFinance); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update fields that are allowed to be changed
		if updatedFinance.AccountTitle != "" {
			finance.AccountTitle = updatedFinance.AccountTitle
		}
		if updatedFinance.InitialBalance != 0 {
			finance.InitialBalance = updatedFinance.InitialBalance
		}
		if updatedFinance.AccountNumber != "" {
			finance.AccountNumber = updatedFinance.AccountNumber
		}
		if updatedFinance.BranchCode != "" {
			finance.BranchCode = updatedFinance.BranchCode
		}
		if updatedFinance.BankBranch != "" {
			finance.BankBranch = updatedFinance.BankBranch
		}

		// Save the updated finance data to the database
		db.Save(&finance)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Finance data updated successfully",
			"data":    finance,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteFinanceByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Extract finance ID from the request
		financeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid finance ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch finance data from the database
		var finance models.Finance
		result = db.First(&finance, uint(financeID))
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Finance data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete finance data from the database
		db.Delete(&finance)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Finance data deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func CreateDepositCategoryByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind deposit category data from request body
		var depositCategory models.DepositCategory
		if err := c.Bind(&depositCategory); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate deposit category data
		if depositCategory.DepositCategory == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Deposit category is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Set creation timestamp
		depositCategory.CreatedAt = time.Now()

		// Create deposit category in the database
		db.Create(&depositCategory)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Deposit category added successfully",
			"data":    depositCategory,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// GetAllDepositCategoriesByAdmin adalah handler untuk mendapatkan semua data deposit category oleh admin
func GetAllDepositCategoriesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Fetch searching query parameter
		searching := c.QueryParam("searching")

		// Fetch deposit categories from database with optional search filter
		var depositCategories []models.DepositCategory
		query := db.Model(&depositCategories)
		if searching != "" {
			query = query.Where("LOWER(deposit_category) LIKE ?", "%"+strings.ToLower(searching)+"%")
		}
		query.Find(&depositCategories)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Deposit categories retrieved successfully",
			"data":    depositCategories,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetDepositCategoryByIDByAdmin adalah handler untuk mendapatkan data deposit category berdasarkan ID oleh admin
func GetDepositCategoryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Fetch deposit category ID from path parameter
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve deposit category from database by ID
		var depositCategory models.DepositCategory
		result = db.First(&depositCategory, id)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Deposit category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Deposit category retrieved successfully",
			"data":    depositCategory,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func EditDepositCategoryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Fetch deposit category ID from path parameter
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Bind the deposit category data from the request body
		var updatedDepositCategory models.DepositCategory
		if err := c.Bind(&updatedDepositCategory); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate deposit category data
		if updatedDepositCategory.DepositCategory == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid deposit category data. Deposit category is required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve deposit category from database by ID
		var depositCategory models.DepositCategory
		result = db.First(&depositCategory, id)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Deposit category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Update deposit category fields
		depositCategory.DepositCategory = updatedDepositCategory.DepositCategory

		// Save the updated deposit category to the database
		db.Save(&depositCategory)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Deposit category updated successfully",
			"data":    depositCategory,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// DeleteDepositCategoryByIDByAdmin adalah handler untuk menghapus data deposit category berdasarkan ID oleh admin
func DeleteDepositCategoryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Fetch deposit category ID from path parameter
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve deposit category from database by ID
		var depositCategory models.DepositCategory
		result = db.First(&depositCategory, id)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Deposit category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the deposit category from the database
		db.Delete(&depositCategory)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Deposit category deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// AddDepositByAdmin adalah handler untuk menambahkan data add deposit baru oleh admin
func AddDepositByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind the add deposit data from the request body
		var addDeposit models.AddDeposit
		if err := c.Bind(&addDeposit); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate deposit data
		if addDeposit.FinanceID == 0 || addDeposit.Amount <= 0 || addDeposit.Date == "" || addDeposit.CategoryID == 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid deposit data. All fields are required and amount must be greater than 0."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate date format
		_, err = time.Parse("2006-01-02", addDeposit.Date)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve finance data from the database
		var finance models.Finance
		result = db.First(&finance, addDeposit.FinanceID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Finance ID not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		addDeposit.AccountTitle = finance.AccountTitle

		// Retrieve deposit category data from the database
		var depositCategory models.DepositCategory
		result = db.First(&depositCategory, addDeposit.CategoryID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Deposit Category ID not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		addDeposit.DepositCategory = depositCategory.DepositCategory

		// Update the initial balance with the deposit amount
		finance.InitialBalance += addDeposit.Amount
		db.Save(&finance)

		// Create the add deposit entry in the database
		db.Create(&addDeposit)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Add deposit data added successfully",
			"data":    addDeposit,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllAddDepositsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Fetch searching query parameter
		searching := c.QueryParam("searching")

		// Fetch add deposit data from database with optional search filters
		var addDeposits []models.AddDeposit
		query := db.Model(&addDeposits)
		if searching != "" {
			query = query.Where("LOWER(account_title) LIKE ? OR amount = ? OR LOWER(date) LIKE ? OR LOWER(deposit_category) LIKE ? OR LOWER(payer) LIKE ? OR LOWER(payment_method) LIKE ? OR LOWER(ref) LIKE ? OR LOWER(description) LIKE ?",
				"%"+strings.ToLower(searching)+"%",
				helper.ParseStringToFloat(searching),
				"%"+strings.ToLower(searching)+"%",
				"%"+strings.ToLower(searching)+"%",
				"%"+strings.ToLower(searching)+"%",
				"%"+strings.ToLower(searching)+"%",
				"%"+strings.ToLower(searching)+"%",
				"%"+strings.ToLower(searching)+"%",
			)
		}
		query.Find(&addDeposits)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Add deposit data retrieved successfully",
			"data":    addDeposits,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetDepositByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Parse deposit ID from path parameter
		depositID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid deposit ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve deposit data from the database
		var deposit models.AddDeposit
		result = db.First(&deposit, depositID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Deposit ID not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Deposit data retrieved successfully",
			"data":    deposit,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateDepositByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Parse deposit ID from path parameter
		depositID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid deposit ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve existing deposit data from the database
		var existingDeposit models.AddDeposit
		result = db.First(&existingDeposit, depositID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Deposit not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated deposit data from the request body
		var updatedDeposit models.AddDeposit
		if err := c.Bind(&updatedDeposit); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve finance data from the database if finance_id is updated
		var newFinance models.Finance
		if updatedDeposit.FinanceID != 0 && updatedDeposit.FinanceID != existingDeposit.FinanceID {
			result = db.First(&newFinance, updatedDeposit.FinanceID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "New finance ID not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
		}

		// Retrieve deposit category data from the database if category_id is updated
		var newDepositCategory models.DepositCategory
		if updatedDeposit.CategoryID != 0 && updatedDeposit.CategoryID != existingDeposit.CategoryID {
			result = db.First(&newDepositCategory, updatedDeposit.CategoryID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "New deposit category ID not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
		}

		// logika untuk mengupdate deposit data ke id finance yang baru
		if updatedDeposit.FinanceID != 0 && updatedDeposit.FinanceID != existingDeposit.FinanceID {
			// menghitung perbedaan amount sebelum dan sesudah
			amountDiff := updatedDeposit.Amount - existingDeposit.Amount

			// mengupdate inital balance untuk finance id baru
			var newFinance models.Finance
			db.First(&newFinance, updatedDeposit.FinanceID)
			newFinance.InitialBalance -= amountDiff
			db.Save(&newFinance)

			// mengupdate initial balance pada finance id yang lama
			var oldFinance models.Finance
			db.First(&oldFinance, existingDeposit.FinanceID)
			oldFinance.InitialBalance += amountDiff
			db.Save(&oldFinance)

			existingDeposit.FinanceID = updatedDeposit.FinanceID
			existingDeposit.AccountTitle = newFinance.AccountTitle
		}

		if updatedDeposit.Amount != 0 {
			// mengitung perbedaan amount
			amountDiff := updatedDeposit.Amount - existingDeposit.Amount
			existingDeposit.Amount = updatedDeposit.Amount

			// mengupdate amount
			var finance models.Finance
			db.First(&finance, existingDeposit.FinanceID)
			finance.InitialBalance += amountDiff
			db.Save(&finance)
		}
		if updatedDeposit.Date != "" {
			existingDeposit.Date = updatedDeposit.Date
		}
		if updatedDeposit.CategoryID != 0 {
			existingDeposit.CategoryID = updatedDeposit.CategoryID
			existingDeposit.DepositCategory = newDepositCategory.DepositCategory
		}
		if updatedDeposit.Payer != "" {
			existingDeposit.Payer = updatedDeposit.Payer
		}
		if updatedDeposit.PaymentMethod != "" {
			existingDeposit.PaymentMethod = updatedDeposit.PaymentMethod
		}
		if updatedDeposit.Ref != "" {
			existingDeposit.Ref = updatedDeposit.Ref
		}
		if updatedDeposit.Description != "" {
			existingDeposit.Description = updatedDeposit.Description
		}

		// Update the database record
		db.Save(&existingDeposit)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Deposit data updated successfully",
			"data":    existingDeposit,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// DeleteDepositByAdmin adalah handler untuk menghapus data add deposit oleh admin berdasarkan ID
func DeleteDepositByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Parse deposit ID from path parameter
		depositID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid deposit ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve existing deposit data from the database
		var existingDeposit models.AddDeposit
		result = db.First(&existingDeposit, depositID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Deposit not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the deposit from the database
		db.Delete(&existingDeposit)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Deposit data deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// AddExpenseCategoryByAdmin adalah handler untuk menambahkan data expense category oleh admin
func CreateExpenseCategoryByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind the expense category data from the request body
		var expenseCategory models.ExpenseCategory
		if err := c.Bind(&expenseCategory); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Create the expense category entry in the database
		db.Create(&expenseCategory)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Expense category added successfully",
			"data":    expenseCategory,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// GetExpenseCategoriesByAdmin adalah handler untuk ADMIN dapat melihat seluruh data expense category dilengkapi dengan fitur searching
func GetAllExpenseCategoriesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Retrieve all expense categories from the database
		var expenseCategories []models.ExpenseCategory
		db.Find(&expenseCategories)

		// Check if searching query parameter is provided
		searching := c.QueryParam("searching")
		if searching != "" {
			search := strings.ToLower(searching)
			var filteredExpenseCategories []models.ExpenseCategory
			for _, category := range expenseCategories {
				if strings.Contains(strings.ToLower(category.ExpenseCategory), search) {
					filteredExpenseCategories = append(filteredExpenseCategories, category)
				}
			}
			expenseCategories = filteredExpenseCategories
		}

		// Respond with the list of expense categories
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Expense categories retrieved successfully",
			"data":    expenseCategories,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetExpenseCategoryByID adalah handler untuk ADMIN dapat melihat data expense category berdasarkan ID
func GetExpenseCategoryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Parse expense category ID from path parameter
		categoryID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid expense category ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve expense category data from the database
		var expenseCategory models.ExpenseCategory
		result = db.First(&expenseCategory, categoryID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with the expense category data
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Expense category retrieved successfully",
			"data":    expenseCategory,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// UpdateExpenseCategoryByID adalah handler untuk ADMIN dapat mengedit data expense category berdasarkan ID
func EditExpenseCategoryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Parse expense category ID from path parameter
		categoryID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid expense category ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve existing expense category data from the database
		var existingCategory models.ExpenseCategory
		result = db.First(&existingCategory, categoryID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated expense category data from the request body
		var updatedCategory models.ExpenseCategory
		if err := c.Bind(&updatedCategory); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update the expense category data with the new values
		if updatedCategory.ExpenseCategory != "" {
			existingCategory.ExpenseCategory = updatedCategory.ExpenseCategory
		}

		// Save the updated expense category data to the database
		db.Save(&existingCategory)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Expense category updated successfully",
			"data":    existingCategory,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// DeleteExpenseCategoryByID adalah handler untuk ADMIN dapat menghapus data expense category berdasarkan ID
func DeleteExpenseCategoryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Parse expense category ID from path parameter
		categoryID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid expense category ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve existing expense category data from the database
		var existingCategory models.ExpenseCategory
		result = db.First(&existingCategory, categoryID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the expense category from the database
		db.Delete(&existingCategory)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Expense category deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func AddExpenseByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind the add expense data from the request body
		var addExpense models.AddExpense
		if err := c.Bind(&addExpense); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate expense data
		if addExpense.FinanceID == 0 || addExpense.Amount <= 0 || addExpense.Date == "" || addExpense.ExpenseCategoryID == 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid expense data. All fields are required and amount must be greater than 0."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate date format
		_, err = time.Parse("2006-01-02", addExpense.Date)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve finance data from the database
		var finance models.Finance
		result = db.First(&finance, addExpense.FinanceID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Finance ID not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		addExpense.AccountTitle = finance.AccountTitle

		// Retrieve expense category data from the database
		var expenseCategory models.ExpenseCategory
		result = db.First(&expenseCategory, addExpense.ExpenseCategoryID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense Category ID not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		addExpense.ExpenseCategory = expenseCategory.ExpenseCategory

		// Update the initial balance with the expense amount
		finance.InitialBalance -= addExpense.Amount
		db.Save(&finance)

		// Create the add expense entry in the database
		db.Create(&addExpense)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Add expense data added successfully",
			"data":    addExpense,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllAddExpensesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Extract search query parameter
		searching := c.QueryParam("searching")

		// Query for expenses with search parameters
		var expenses []models.AddExpense
		query := db.Model(&expenses)
		if searching != "" {
			searching = strings.ToLower(searching)
			query = query.Where("LOWER(account_title) LIKE ? OR amount = ? OR LOWER(date) LIKE ? OR LOWER(expense_category) LIKE ? OR LOWER(payer) LIKE ? OR LOWER(payment_method) LIKE ? OR LOWER(ref) LIKE ? OR LOWER(description) LIKE ?",
				"%"+searching+"%",
				helper.ParseStringToFloat(searching),
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
			)
		}
		query.Find(&expenses)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Expenses retrieved successfully",
			"data":    expenses,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetExpenseByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Parse expense ID from path parameter
		expenseID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid expense ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve expense data from the database
		var expense models.AddExpense
		result = db.First(&expense, expenseID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Expense retrieved successfully",
			"data":    expense,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateExpenseByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Parse expense ID from path parameter
		expenseID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid expense ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve existing expense data from the database
		var existingExpense models.AddExpense
		result = db.First(&existingExpense, expenseID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated expense data from the request body
		var updatedExpense models.AddExpense
		if err := c.Bind(&updatedExpense); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve finance data from the database if finance_id is updated
		var newFinance models.Finance
		if updatedExpense.FinanceID != 0 && updatedExpense.FinanceID != existingExpense.FinanceID {
			result = db.First(&newFinance, updatedExpense.FinanceID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "New finance ID not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}

			// Calculate the difference in amount
			amountDiff := updatedExpense.Amount - existingExpense.Amount

			// Update the initial balance of the new finance ID
			newFinance.InitialBalance += amountDiff
			db.Save(&newFinance)

			// Update the initial balance of the old finance ID
			var oldFinance models.Finance
			db.First(&oldFinance, existingExpense.FinanceID)
			oldFinance.InitialBalance -= amountDiff
			db.Save(&oldFinance)

			// Update the finance ID and account title in the expense data
			existingExpense.FinanceID = updatedExpense.FinanceID
			existingExpense.AccountTitle = newFinance.AccountTitle
		}

		if updatedExpense.Amount != 0 {
			// Calculate the difference in amount
			amountDiff := updatedExpense.Amount - existingExpense.Amount
			existingExpense.Amount = updatedExpense.Amount

			// Update the initial balance of finance ID
			var finance models.Finance
			db.First(&finance, existingExpense.FinanceID)
			finance.InitialBalance -= amountDiff
			db.Save(&finance)
		}
		if updatedExpense.Date != "" {
			existingExpense.Date = updatedExpense.Date
		}
		if updatedExpense.ExpenseCategoryID != 0 {
			existingExpense.ExpenseCategoryID = updatedExpense.ExpenseCategoryID

			// Retrieve expense category data from the database
			var expenseCategory models.ExpenseCategory
			result = db.First(&expenseCategory, updatedExpense.ExpenseCategoryID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense Category ID not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}

			existingExpense.ExpenseCategory = expenseCategory.ExpenseCategory
		}
		if updatedExpense.Payer != "" {
			existingExpense.Payer = updatedExpense.Payer
		}
		if updatedExpense.PaymentMethod != "" {
			existingExpense.PaymentMethod = updatedExpense.PaymentMethod
		}
		if updatedExpense.Ref != "" {
			existingExpense.Ref = updatedExpense.Ref
		}
		if updatedExpense.Description != "" {
			existingExpense.Description = updatedExpense.Description
		}

		// Update the database record
		db.Save(&existingExpense)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Expense data updated successfully",
			"data":    existingExpense,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteExpenseByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Parse expense ID from path parameter
		expenseID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid expense ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve existing expense data from the database
		var expense models.AddExpense
		result = db.First(&expense, expenseID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Retrieve finance data from the database
		var finance models.Finance
		result = db.First(&finance, expense.FinanceID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Finance ID not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Update the initial balance by subtracting the expense amount
		finance.InitialBalance += expense.Amount
		db.Save(&finance)

		// Delete the expense record from the database
		db.Delete(&expense)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Expense data deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllTransactions(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Fetch add deposit data from database with preloading DepositCategory
		var addDeposits []models.AddDeposit
		db.Preload("DepositCategory").Find(&addDeposits)

		// Fetch add expense data from database with preloading ExpenseCategory
		var addExpenses []models.AddExpense
		db.Preload("ExpenseCategory").Find(&addExpenses)

		// Combine both results
		transactions := append([]models.AddDeposit{}, addDeposits...)
		for _, expense := range addExpenses {
			// Convert expense to AddDeposit type
			addDeposit := models.AddDeposit{
				ID:              expense.ID,
				FinanceID:       expense.FinanceID,
				AccountTitle:    expense.AccountTitle,
				Amount:          expense.Amount,
				Date:            expense.Date,
				CategoryID:      expense.ExpenseCategoryID,
				DepositCategory: expense.ExpenseCategory,
				Payer:           expense.Payer,
				PaymentMethod:   expense.PaymentMethod,
				Ref:             expense.Ref,
				Description:     expense.Description,
				CreatedAt:       expense.CreatedAt,
			}
			transactions = append(transactions, addDeposit)
		}

		// Sort transactions by createdAt in descending order
		sort.Slice(transactions, func(i, j int) bool {
			return transactions[i].CreatedAt.After(transactions[j].CreatedAt)
		})

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Transactions data retrieved successfully",
			"data":    transactions,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
