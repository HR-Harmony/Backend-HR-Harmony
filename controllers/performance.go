package controllers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"strconv"
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

// GetAllGoalTypesByAdmin retrieves all goal types information with pagination
func GetAllGoalTypesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Parse the token to get admin's username
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

		// Pagination parameters
		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10 // Default per page
		}

		// Calculate offset and limit for pagination
		offset := (page - 1) * perPage

		// Fetch all goal types with pagination
		var goalTypes []models.GoalType
		db.Offset(offset).Limit(perPage).Find(&goalTypes)

		// Get total count of goal types
		var totalCount int64
		db.Model(&models.GoalType{}).Count(&totalCount)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "All goal types retrieved successfully",
			"goalTypes":  goalTypes,
			"pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
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

		goal.GoalRating = 0
		goal.ProgressBar = 0
		goal.Status = "Not Started"

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

// GetAllGoalsByAdmin retrieves all goals information with pagination
func GetAllGoalsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Parse the token to get admin's username
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

		// Pagination parameters
		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10 // Default per page
		}

		// Calculate offset and limit for pagination
		offset := (page - 1) * perPage

		// Fetch all goals with pagination
		var goals []models.Goal
		db.Offset(offset).Limit(perPage).Find(&goals)

		// Get total count of goals
		var totalCount int64
		db.Model(&models.Goal{}).Count(&totalCount)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "All goals retrieved successfully",
			"goals":      goals,
			"pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
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

		if updatedGoal.ProjectID != 0 {
			goal.ProjectID = updatedGoal.ProjectID
			var project models.Project
			result := db.First(&project, updatedGoal.ProjectID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Project not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			goal.ProjectName = project.Title
		}

		if updatedGoal.TaskID != 0 {
			goal.TaskID = updatedGoal.TaskID
			var task models.Task
			result := db.First(&task, updatedGoal.TaskID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Task not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			goal.TaskName = task.Title
		}

		if updatedGoal.TrainingID != 0 {
			goal.TrainingID = updatedGoal.TrainingID
			var training models.Training
			result := db.First(&training, updatedGoal.TrainingID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Training not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
		}

		if updatedGoal.TrainingSkillID != 0 {
			goal.TrainingSkillID = updatedGoal.TrainingSkillID
			var trainingSkill models.TrainingSkill
			result := db.First(&trainingSkill, updatedGoal.TrainingSkillID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Training Skill ID not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			goal.TrainingSkillName = trainingSkill.TrainingSkill
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

		if updatedGoal.GoalRating != 0 {
			goal.GoalRating = updatedGoal.GoalRating
		}

		if updatedGoal.GoalRating < 0 || updatedGoal.GoalRating > 5 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid GoalRating. Must be between 0 and 5."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedGoal.ProgressBar != 0 {
			goal.ProgressBar = updatedGoal.ProgressBar
		}

		if updatedGoal.Status != "" {
			goal.Status = updatedGoal.Status
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
		kpiIndicator.Result = totalScores / 37

		// Set admin ID and username
		kpiIndicator.AdminId = adminUser.ID
		kpiIndicator.AdminName = adminUser.FirstName + " " + adminUser.LastName

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

// GetAllKPIIndicatorsByAdmin retrieves all KPI indicators with pagination
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

		// Pagination parameters
		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10 // Default per page
		}

		// Calculate offset and limit for pagination
		offset := (page - 1) * perPage

		// Fetch KPI indicators with pagination
		var kpiIndicators []models.KPIIndicator
		db.Offset(offset).Limit(perPage).Find(&kpiIndicators)

		// Get total count of KPI indicators
		var totalCount int64
		db.Model(&models.KPIIndicator{}).Count(&totalCount)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "KPI indicators fetched successfully",
			"data":       kpiIndicators,
			"pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetKPIIndicatorsByIdByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve performance ID from the URL parameter
		performanceIDStr := c.Param("id")
		performanceID, err := strconv.ParseUint(performanceIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid KPI Indicator ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the performance from the database
		var kpiIndicator models.KPIIndicator
		result = db.First(&kpiIndicator, uint(performanceID))
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "KPI Indicator not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "KPI Indicator retrieved successfully",
			"data":    kpiIndicator,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func EditKPIIndicatorByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve KPI Indicator ID from the URL parameter
		kpiIndicatorIDStr := c.Param("id")
		kpiIndicatorID, err := strconv.ParseUint(kpiIndicatorIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid KPI Indicator ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the KPI Indicator from the database
		var kpiIndicator models.KPIIndicator
		result = db.First(&kpiIndicator, uint(kpiIndicatorID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "KPI Indicator not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated KPI Indicator data from the request body
		var updateData models.KPIIndicator
		if err := c.Bind(&updateData); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate and update KPI Indicator data
		if updateData.DesignationID != 0 {
			// Retrieve the designation name based on the new designation ID
			var newDesignation models.Designation
			result = db.First(&newDesignation, updateData.DesignationID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Designation not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			kpiIndicator.DesignationName = newDesignation.DesignationName
		}

		// Update only the provided fields
		db.Model(&kpiIndicator).Updates(updateData)

		// Recalculate the result based on the updated scores
		totalScores := helper.CalculateTotalScores(kpiIndicator)
		kpiIndicator.Result = totalScores / 37

		// Save the updated KPI Indicator
		db.Save(&kpiIndicator)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "KPI Indicator updated successfully",
			"data":    kpiIndicator,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteKPIIndicatorByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve KPI Indicator ID from the URL parameter
		kpiIndicatorIDStr := c.Param("id")
		kpiIndicatorID, err := strconv.ParseUint(kpiIndicatorIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid KPI Indicator ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the KPI Indicator from the database
		var kpiIndicator models.KPIIndicator
		result = db.First(&kpiIndicator, uint(kpiIndicatorID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "KPI Indicator not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the KPI Indicator
		db.Delete(&kpiIndicator)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "KPI Indicator deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// KPA Indicator

func CreateKPAIndicatorByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Bind the KPA Indicator data from the request body
		var kpaIndicator models.KPAIndicator
		if err := c.Bind(&kpaIndicator); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate KPA Indicator data
		// Check if designation_id is provided
		if kpaIndicator.EmployeeID == 0 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Employee ID is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the designation name based on designation_id
		var employee models.Employee
		result = db.First(&employee, kpaIndicator.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		kpaIndicator.EmployeeName = employee.FirstName + " " + employee.LastName

		// Validate KPA scores
		if !helper.IsValidScoreKPA(kpaIndicator) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid score. Scores should be between 0 and 5"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Calculate the result based on technical and organizational scores
		totalScores := helper.CalculateTotalScoresKPA(kpaIndicator)
		kpaIndicator.Result = totalScores / 37

		// Set admin ID and username
		kpaIndicator.AdminId = adminUser.ID
		kpaIndicator.AdminName = adminUser.FirstName + " " + adminUser.LastName

		// Validate appraisal_date format
		if kpaIndicator.AppraisalDate == "" || !helper.IsValidAppraisalDateFormat(kpaIndicator.AppraisalDate) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid appraisal date format. Please use mm-yyyy format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Convert appraisal_date to time.Time
		appraisalTime, err := time.Parse("01-2006", kpaIndicator.AppraisalDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid appraisal date format. Please use mm-yyyy format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		kpaIndicator.AppraisalDate = appraisalTime.Format("2006-01")

		// Create the KPI Indicator in the database
		db.Create(&kpaIndicator)

		// Kirim notifikasi email ke karyawan
		err = helper.SendKPAAppraisalNotification(employee.Email, kpaIndicator.AppraisalDate, kpaIndicator.Result)
		if err != nil {
			// Handle error
			fmt.Println("Failed to send KPA appraisal notification:", err)
		}

		successResponse := map[string]interface{}{
			"code":          http.StatusCreated,
			"error":         false,
			"message":       "KPA Indicator created successfully",
			"kpa_indicator": &kpaIndicator,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// GetAllKPAIndicatorsByAdmin retrieves all KPA indicators with pagination
func GetAllKPAIndicatorsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Pagination parameters
		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10 // Default per page
		}

		// Calculate offset and limit for pagination
		offset := (page - 1) * perPage

		// Fetch KPA indicators with pagination
		var kpaIndicators []models.KPAIndicator
		db.Offset(offset).Limit(perPage).Find(&kpaIndicators)

		// Get total count of KPA indicators
		var totalCount int64
		db.Model(&models.KPAIndicator{}).Count(&totalCount)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "KPA indicators fetched successfully",
			"data":       kpaIndicators,
			"pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetKPAIndicatorsByIdByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve performance ID from the URL parameter
		performanceIDStr := c.Param("id")
		performanceID, err := strconv.ParseUint(performanceIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid KPA Indicator ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the performance from the database
		var kpaIndicator models.KPAIndicator
		result = db.First(&kpaIndicator, uint(performanceID))
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "KPA Indicator not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "KPA Indicator retrieved successfully",
			"data":    kpaIndicator,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func EditKPAIndicatorByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve KPA Indicator ID from the URL parameter
		kpaIndicatorIDStr := c.Param("id")
		kpaIndicatorID, err := strconv.ParseUint(kpaIndicatorIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid KPA Indicator ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the KPA Indicator from the database
		var kpaIndicator models.KPAIndicator
		result = db.First(&kpaIndicator, uint(kpaIndicatorID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "KPA Indicator not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated KPA Indicator data from the request body
		var updateData models.KPAIndicator
		if err := c.Bind(&updateData); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate and update KPA Indicator data
		if updateData.EmployeeID != 0 {
			// Retrieve the designation name based on the new designation ID
			var newEmployee models.Employee
			result = db.First(&newEmployee, updateData.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			kpaIndicator.EmployeeName = newEmployee.FirstName + " " + newEmployee.LastName
		}

		// Update only the provided fields
		db.Model(&kpaIndicator).Updates(updateData)

		// Recalculate the result based on the updated scores
		totalScores := helper.CalculateTotalScoresKPA(kpaIndicator)
		kpaIndicator.Result = totalScores / 37

		// Save the updated KPA Indicator
		db.Save(&kpaIndicator)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "KPA Indicator updated successfully",
			"data":    kpaIndicator,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteKPAIndicatorByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve KPI Indicator ID from the URL parameter
		kpaIndicatorIDStr := c.Param("id")
		kpaIndicatorID, err := strconv.ParseUint(kpaIndicatorIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid KPA Indicator ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the KPI Indicator from the database
		var kpaIndicator models.KPAIndicator
		result = db.First(&kpaIndicator, uint(kpaIndicatorID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "KPA Indicator not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the KPI Indicator
		db.Delete(&kpaIndicator)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "KPA Indicator deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
