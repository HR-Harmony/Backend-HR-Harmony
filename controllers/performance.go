package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"strings"
)

// Goal Type

func CreateGoalTypeByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Mendapatkan token dari header Authorization
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Memeriksa format token
		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mendapatkan nilai token
		tokenString = authParts[1]

		// Verifikasi token
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mencari admin berdasarkan username
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Memeriksa apakah pengguna adalah admin HR
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind data goal type dari request
		var goalType models.GoalType
		if err := c.Bind(&goalType); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validasi data goal type
		if goalType.GoalType == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal type cannot be empty"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Membuat data goal type
		db.Create(&goalType)
		// Response sukses
		response := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Goal type added successfully",
			"data":    goalType,
		}
		return c.JSON(http.StatusOK, response)
	}
}

func GetAllGoalTypesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Mendapatkan token dari header Authorization
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Memeriksa format token
		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mendapatkan nilai token
		tokenString = authParts[1]

		// Verifikasi token
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mencari admin berdasarkan username
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Memeriksa apakah pengguna adalah admin HR
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Mendapatkan semua data goal type
		var goalTypes []models.GoalType
		db.Find(&goalTypes)

		// Response sukses
		response := map[string]interface{}{
			"code":      http.StatusOK,
			"error":     false,
			"message":   "All goal types retrieved successfully",
			"goalTypes": goalTypes,
		}
		return c.JSON(http.StatusOK, response)
	}
}

func GetGoalTypeByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Mendapatkan token dari header Authorization
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Memeriksa format token
		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mendapatkan nilai token
		tokenString = authParts[1]

		// Verifikasi token
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mencari admin berdasarkan username
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Memeriksa apakah pengguna adalah admin HR
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Mendapatkan parameter ID dari URL
		goalTypeID := c.Param("id")
		if goalTypeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal type ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Mencari goal type berdasarkan ID
		var goalType models.GoalType
		result = db.First(&goalType, "id = ?", goalTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Response sukses
		response := map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"message":  "Goal type retrieved successfully",
			"goalType": goalType,
		}
		return c.JSON(http.StatusOK, response)
	}
}

func UpdateGoalTypeByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Mendapatkan token dari header Authorization
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Memeriksa format token
		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mendapatkan nilai token
		tokenString = authParts[1]

		// Verifikasi token
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mencari admin berdasarkan username
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Memeriksa apakah pengguna adalah admin HR
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Mendapatkan parameter ID dari URL
		goalTypeID := c.Param("id")
		if goalTypeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal type ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Mencari goal type berdasarkan ID
		var goalType models.GoalType
		result = db.First(&goalType, "id = ?", goalTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind data goal type dari request
		var updatedGoalType models.GoalType
		if err := c.Bind(&updatedGoalType); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validasi data goal type
		if updatedGoalType.GoalType == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal type cannot be empty"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update data goal type
		goalType.GoalType = updatedGoalType.GoalType
		goalType.CreatedAt = updatedGoalType.CreatedAt

		db.Save(&goalType)

		// Response sukses
		response := map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"message":  "Goal type updated successfully",
			"goalType": goalType,
		}
		return c.JSON(http.StatusOK, response)
	}
}

func DeleteGoalTypeByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Mendapatkan token dari header Authorization
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Memeriksa format token
		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mendapatkan nilai token
		tokenString = authParts[1]

		// Verifikasi token
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mencari admin berdasarkan username
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Memeriksa apakah pengguna adalah admin HR
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Mendapatkan parameter ID dari URL
		goalTypeID := c.Param("id")
		if goalTypeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal type ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Mencari goal type berdasarkan ID
		var goalType models.GoalType
		result = db.First(&goalType, "id = ?", goalTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Hapus goal type
		db.Delete(&goalType)

		// Response sukses
		response := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Goal type deleted successfully",
		}
		return c.JSON(http.StatusOK, response)
	}
}
