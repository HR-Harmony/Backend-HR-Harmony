package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"strings"
	"time"
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

// Tracking Goals

func CreateGoalByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Bind data goal dari request
		var goal models.Goal
		if err := c.Bind(&goal); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validasi data goal
		if goal.GoalTypeID == 0 || goal.Subject == "" || goal.TargetAchievement == "" || goal.StartDate == "" || goal.EndDate == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid goal data. All fields are required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Set goal_type_name berdasarkan goal_type_id
		var goalType models.GoalType
		result = db.First(&goalType, "id = ?", goal.GoalTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		goal.GoalTypeName = goalType.GoalType

		// Parse start date from string to time.Time
		if goal.StartDate != "" {
			startDate, err := time.Parse("2006-01-02", goal.StartDate)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid StartDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			// Format start date in "yyyy-mm-dd" format
			goal.StartDate = startDate.Format("2006-01-02")
		}

		// Parse end date from string to time.Time
		if goal.EndDate != "" {
			endDate, err := time.Parse("2006-01-02", goal.EndDate)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid EndDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			// Format end date in "yyyy-mm-dd" format
			goal.EndDate = endDate.Format("2006-01-02")
		}

		// Set created_at
		goal.CreatedAt = time.Now().Format("2006-01-02")

		// Membuat data goal
		db.Create(&goal)

		// Response sukses
		response := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Goal added successfully",
			"data":    goal,
		}
		return c.JSON(http.StatusOK, response)
	}
}

func GetAllGoalsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Mencari semua data tracking goals
		var goals []models.Goal
		db.Find(&goals)

		// Response sukses
		response := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "All goals retrieved successfully",
			"goals":   goals,
		}
		return c.JSON(http.StatusOK, response)
	}
}

func GetGoalByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Mendapatkan ID goal dari parameter URL
		goalID := c.Param("id")
		if goalID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Mencari data tracking goal berdasarkan ID
		var goal models.Goal
		result = db.First(&goal, "id = ?", goalID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Response sukses
		response := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Goal retrieved successfully",
			"data":    goal,
		}
		return c.JSON(http.StatusOK, response)
	}
}

func UpdateGoalByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Mendapatkan ID goal dari parameter URL
		goalID := c.Param("id")
		if goalID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Mencari data tracking goal berdasarkan ID
		var goal models.Goal
		result = db.First(&goal, "id = ?", goalID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind data goal dari request
		var updatedGoal models.Goal
		if err := c.Bind(&updatedGoal); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update field yang diizinkan diubah
		if updatedGoal.GoalTypeID != 0 {
			var goalType models.GoalType
			result = db.First(&goalType, "id = ?", updatedGoal.GoalTypeID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid goal type ID. Goal type not found."}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			updatedGoal.GoalTypeID = goalType.ID
			updatedGoal.GoalTypeName = goalType.GoalType

			// Set nilai pada goal
			goal.GoalTypeID = updatedGoal.GoalTypeID
			goal.GoalTypeName = updatedGoal.GoalTypeName
		}

		if updatedGoal.Subject != "" {
			goal.Subject = updatedGoal.Subject
		}
		if updatedGoal.TargetAchievement != "" {
			goal.TargetAchievement = updatedGoal.TargetAchievement
		}
		if updatedGoal.StartDate != "" {
			startDate, err := time.Parse("2006-01-02", updatedGoal.StartDate)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid StartDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			goal.StartDate = startDate.Format("2006-01-02")
		}
		if updatedGoal.EndDate != "" {
			endDate, err := time.Parse("2006-01-02", updatedGoal.EndDate)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid EndDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			goal.EndDate = endDate.Format("2006-01-02")
		}
		if updatedGoal.Description != "" {
			goal.Description = updatedGoal.Description
		}

		// Simpan perubahan ke database
		db.Save(&goal)

		// Response sukses
		response := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Goal updated successfully",
			"data":    goal,
		}
		return c.JSON(http.StatusOK, response)
	}
}

func DeleteGoalByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Mendapatkan ID goal dari parameter URL
		goalID := c.Param("id")
		if goalID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Mencari data tracking goal berdasarkan ID
		var goal models.Goal
		result = db.First(&goal, "id = ?", goalID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Menghapus data goal dari database
		db.Delete(&goal)

		// Response sukses
		response := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Goal deleted successfully",
		}
		return c.JSON(http.StatusOK, response)
	}
}

func CreateKPIIndicatorByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind the KPI Indicator data from the request body
		var kpiIndicator models.KPIIndicator
		if err := c.Bind(&kpiIndicator); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate KPI Indicator data
		// Check if designation_id is provided
		if kpiIndicator.DesignationID == 0 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Designation ID is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the designation name based on designation_id
		var designation models.Designation
		result = db.First(&designation, kpiIndicator.DesignationID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Designation not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		kpiIndicator.DesignationName = designation.DesignationName

		// Validate KPI scores
		if !helper.IsValidScore(kpiIndicator) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid score. Scores should be between 0 and 5"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Calculate the result based on technical and organizational scores
		totalScores := helper.CalculateTotalScores(kpiIndicator)
		kpiIndicator.Result = totalScores / 36

		// Create the KPI Indicator in the database
		db.Create(&kpiIndicator)

		successResponse := map[string]interface{}{
			"code":          http.StatusCreated,
			"error":         false,
			"message":       "KPI Indicator created successfully",
			"kpi_indicator": &kpiIndicator,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllKPIIndicatorsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Fetch all KPI Indicators from the database
		var kpiIndicators []models.KPIIndicator
		if err := db.Find(&kpiIndicators).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch KPI Indicators"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "KPI Indicators fetched successfully",
			"data":    kpiIndicators,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
