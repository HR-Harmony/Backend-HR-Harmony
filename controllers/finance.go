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

		var finance models.Finance
		if err := c.Bind(&finance); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if finance.AccountTitle == "" || finance.InitialBalance == 0 || finance.AccountNumber == "" || finance.BranchCode == "" || finance.BankBranch == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid finance data. All fields are required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		db.Create(&finance)

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

		searching := c.QueryParam("searching")

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		var finances []models.Finance
		query := db.Model(&finances)
		if searching != "" {
			query = query.Where("LOWER(account_title) ILIKE ? OR cast(initial_balance as text) LIKE ? OR account_number LIKE ? OR branch_code LIKE ? OR LOWER(bank_branch) ILIKE ?", "%"+strings.ToLower(searching)+"%", "%"+searching+"%", "%"+searching+"%", "%"+searching+"%", "%"+strings.ToLower(searching)+"%")
		}

		var totalCount int64
		query.Count(&totalCount)

		offset := (page - 1) * perPage

		query.Offset(offset).Limit(perPage).Find(&finances)

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

		financeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid finance ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var finance models.Finance
		result = db.First(&finance, uint(financeID))
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Finance data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

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

		financeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid finance ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var finance models.Finance
		result = db.First(&finance, uint(financeID))
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Finance data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedFinance models.Finance
		if err := c.Bind(&updatedFinance); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

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

		db.Save(&finance)

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

		financeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid finance ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var finance models.Finance
		result = db.First(&finance, uint(financeID))
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Finance data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&finance)

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

		var depositCategory models.DepositCategory
		if err := c.Bind(&depositCategory); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if depositCategory.DepositCategory == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Deposit category is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		depositCategory.CreatedAt = time.Now()

		db.Create(&depositCategory)

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Deposit category added successfully",
			"data":    depositCategory,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllDepositCategoriesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		searching := c.QueryParam("searching")

		var depositCategories []models.DepositCategory
		query := db.Model(&depositCategories)
		if searching != "" {
			query = query.Where("LOWER(deposit_category) LIKE ?", "%"+strings.ToLower(searching)+"%")
		}
		query.Find(&depositCategories)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Deposit categories retrieved successfully",
			"data":    depositCategories,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetDepositCategoryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var depositCategory models.DepositCategory
		result = db.First(&depositCategory, id)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Deposit category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

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

		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var updatedDepositCategory models.DepositCategory
		if err := c.Bind(&updatedDepositCategory); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedDepositCategory.DepositCategory == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid deposit category data. Deposit category is required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var depositCategory models.DepositCategory
		result = db.First(&depositCategory, id)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Deposit category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		depositCategory.DepositCategory = updatedDepositCategory.DepositCategory

		db.Save(&depositCategory)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Deposit category updated successfully",
			"data":    depositCategory,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteDepositCategoryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var depositCategory models.DepositCategory
		result = db.First(&depositCategory, id)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Deposit category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&depositCategory)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Deposit category deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func AddDepositByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		var addDeposit models.AddDeposit
		if err := c.Bind(&addDeposit); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if addDeposit.FinanceID == 0 || addDeposit.Amount <= 0 || addDeposit.Date == "" || addDeposit.CategoryID == 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid deposit data. All fields are required and amount must be greater than 0."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		_, err = time.Parse("2006-01-02", addDeposit.Date)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var finance models.Finance
		result = db.First(&finance, addDeposit.FinanceID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Finance ID not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		addDeposit.AccountTitle = finance.AccountTitle

		var depositCategory models.DepositCategory
		result = db.First(&depositCategory, addDeposit.CategoryID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Deposit Category ID not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		addDeposit.DepositCategory = depositCategory.DepositCategory

		finance.InitialBalance += addDeposit.Amount
		db.Save(&finance)

		db.Create(&addDeposit)

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

		searching := c.QueryParam("searching")

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

		depositID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid deposit ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var deposit models.AddDeposit
		result = db.First(&deposit, depositID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Deposit ID not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

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

		depositID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid deposit ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingDeposit models.AddDeposit
		result = db.First(&existingDeposit, depositID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Deposit not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedDeposit models.AddDeposit
		if err := c.Bind(&updatedDeposit); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var newFinance models.Finance
		if updatedDeposit.FinanceID != 0 && updatedDeposit.FinanceID != existingDeposit.FinanceID {
			result = db.First(&newFinance, updatedDeposit.FinanceID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "New finance ID not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
		}

		var newDepositCategory models.DepositCategory
		if updatedDeposit.CategoryID != 0 && updatedDeposit.CategoryID != existingDeposit.CategoryID {
			result = db.First(&newDepositCategory, updatedDeposit.CategoryID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "New deposit category ID not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
		}

		if updatedDeposit.FinanceID != 0 && updatedDeposit.FinanceID != existingDeposit.FinanceID {
			amountDiff := updatedDeposit.Amount - existingDeposit.Amount

			var newFinance models.Finance
			db.First(&newFinance, updatedDeposit.FinanceID)
			newFinance.InitialBalance -= amountDiff
			db.Save(&newFinance)

			var oldFinance models.Finance
			db.First(&oldFinance, existingDeposit.FinanceID)
			oldFinance.InitialBalance += amountDiff
			db.Save(&oldFinance)

			existingDeposit.FinanceID = updatedDeposit.FinanceID
			existingDeposit.AccountTitle = newFinance.AccountTitle
		}

		if updatedDeposit.Amount != 0 {
			amountDiff := updatedDeposit.Amount - existingDeposit.Amount
			existingDeposit.Amount = updatedDeposit.Amount

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

		db.Save(&existingDeposit)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Deposit data updated successfully",
			"data":    existingDeposit,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteDepositByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		depositID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid deposit ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingDeposit models.AddDeposit
		result = db.First(&existingDeposit, depositID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Deposit not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&existingDeposit)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Deposit data deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func CreateExpenseCategoryByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		var expenseCategory models.ExpenseCategory
		if err := c.Bind(&expenseCategory); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		db.Create(&expenseCategory)

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Expense category added successfully",
			"data":    expenseCategory,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllExpenseCategoriesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		var expenseCategories []models.ExpenseCategory
		db.Find(&expenseCategories)

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

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Expense categories retrieved successfully",
			"data":    expenseCategories,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetExpenseCategoryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		categoryID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid expense category ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var expenseCategory models.ExpenseCategory
		result = db.First(&expenseCategory, categoryID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Expense category retrieved successfully",
			"data":    expenseCategory,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func EditExpenseCategoryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		categoryID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid expense category ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingCategory models.ExpenseCategory
		result = db.First(&existingCategory, categoryID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedCategory models.ExpenseCategory
		if err := c.Bind(&updatedCategory); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedCategory.ExpenseCategory != "" {
			existingCategory.ExpenseCategory = updatedCategory.ExpenseCategory
		}

		db.Save(&existingCategory)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Expense category updated successfully",
			"data":    existingCategory,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteExpenseCategoryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		categoryID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid expense category ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingCategory models.ExpenseCategory
		result = db.First(&existingCategory, categoryID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&existingCategory)

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

		var addExpense models.AddExpense
		if err := c.Bind(&addExpense); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if addExpense.FinanceID == 0 || addExpense.Amount <= 0 || addExpense.Date == "" || addExpense.ExpenseCategoryID == 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid expense data. All fields are required and amount must be greater than 0."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		_, err = time.Parse("2006-01-02", addExpense.Date)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var finance models.Finance
		result = db.First(&finance, addExpense.FinanceID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Finance ID not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		addExpense.AccountTitle = finance.AccountTitle

		var expenseCategory models.ExpenseCategory
		result = db.First(&expenseCategory, addExpense.ExpenseCategoryID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense Category ID not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		addExpense.ExpenseCategory = expenseCategory.ExpenseCategory

		finance.InitialBalance -= addExpense.Amount
		db.Save(&finance)

		db.Create(&addExpense)

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

		searching := c.QueryParam("searching")

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

		expenseID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid expense ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var expense models.AddExpense
		result = db.First(&expense, expenseID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

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

		expenseID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid expense ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingExpense models.AddExpense
		result = db.First(&existingExpense, expenseID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedExpense models.AddExpense
		if err := c.Bind(&updatedExpense); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var newFinance models.Finance
		if updatedExpense.FinanceID != 0 && updatedExpense.FinanceID != existingExpense.FinanceID {
			result = db.First(&newFinance, updatedExpense.FinanceID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "New finance ID not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}

			amountDiff := updatedExpense.Amount - existingExpense.Amount

			newFinance.InitialBalance += amountDiff
			db.Save(&newFinance)

			var oldFinance models.Finance
			db.First(&oldFinance, existingExpense.FinanceID)
			oldFinance.InitialBalance -= amountDiff
			db.Save(&oldFinance)

			existingExpense.FinanceID = updatedExpense.FinanceID
			existingExpense.AccountTitle = newFinance.AccountTitle
		}

		if updatedExpense.Amount != 0 {
			amountDiff := updatedExpense.Amount - existingExpense.Amount
			existingExpense.Amount = updatedExpense.Amount

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

		db.Save(&existingExpense)

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

		expenseID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid expense ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var expense models.AddExpense
		result = db.First(&expense, expenseID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Expense not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var finance models.Finance
		result = db.First(&finance, expense.FinanceID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Finance ID not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		finance.InitialBalance += expense.Amount
		db.Save(&finance)

		db.Delete(&expense)

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

		var addDeposits []models.AddDeposit
		db.Preload("DepositCategory").Find(&addDeposits)

		var addExpenses []models.AddExpense
		db.Preload("ExpenseCategory").Find(&addExpenses)

		transactions := append([]models.AddDeposit{}, addDeposits...)
		for _, expense := range addExpenses {
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

		sort.Slice(transactions, func(i, j int) bool {
			return transactions[i].CreatedAt.After(transactions[j].CreatedAt)
		})

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Transactions data retrieved successfully",
			"data":    transactions,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
