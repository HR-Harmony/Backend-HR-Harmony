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

		var goalType models.GoalType
		if err := c.Bind(&goalType); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if goalType.GoalType == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal type cannot be empty"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		db.Create(&goalType)
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

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		searching := c.QueryParam("searching")

		query := db.Model(&models.GoalType{})
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where("LOWER(goal_type) LIKE ?", searchPattern)
		}

		var totalCount int64
		query.Count(&totalCount)

		var goalTypes []models.GoalType
		if err := query.Offset(offset).Limit(perPage).Find(&goalTypes).Order("id DESC").Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching goal types"})
		}

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

		goalTypeID := c.Param("id")
		if goalTypeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal type ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var goalType models.GoalType
		result = db.First(&goalType, "id = ?", goalTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

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

		goalTypeID := c.Param("id")
		if goalTypeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal type ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var goalType models.GoalType
		result = db.First(&goalType, "id = ?", goalTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedGoalType models.GoalType
		if err := c.Bind(&updatedGoalType); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedGoalType.GoalType == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal type cannot be empty"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		goalType.GoalType = updatedGoalType.GoalType
		goalType.CreatedAt = updatedGoalType.CreatedAt

		db.Save(&goalType)

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

		goalTypeID := c.Param("id")
		if goalTypeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal type ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var goalType models.GoalType
		result = db.First(&goalType, "id = ?", goalTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&goalType)

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

		var goal models.Goal
		if err := c.Bind(&goal); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if goal.GoalTypeID == 0 || goal.Subject == "" || goal.TargetAchievement == "" || goal.StartDate == "" || goal.EndDate == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid goal data. All fields are required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var goalType models.GoalType
		result = db.First(&goalType, "id = ?", goal.GoalTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		goal.GoalTypeName = goalType.GoalType

		if goal.StartDate != "" {
			startDate, err := time.Parse("2006-01-02", goal.StartDate)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid StartDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			goal.StartDate = startDate.Format("2006-01-02")
		}

		if goal.EndDate != "" {
			endDate, err := time.Parse("2006-01-02", goal.EndDate)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid EndDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			goal.EndDate = endDate.Format("2006-01-02")
		}

		goal.GoalRating = 0
		goal.ProgressBar = 0
		goal.Status = "Not Started"

		goal.CreatedAt = time.Now().Format("2006-01-02")

		db.Create(&goal)

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

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		searching := c.QueryParam("searching")

		query := db.Model(&models.Goal{})
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where(
				"LOWER(goal_type_name) LIKE ? OR LOWER(subject) LIKE ? OR LOWER(start_date) LIKE ? OR LOWER(end_date) LIKE ? OR LOWER(status) LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
			)
		}

		var totalCount int64
		query.Count(&totalCount)

		var goals []models.Goal
		if err := query.Offset(offset).Limit(perPage).Find(&goals).Order("id DESC").Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching goals"})
		}

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

		goalID := c.Param("id")
		if goalID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var goal models.Goal
		result = db.First(&goal, "id = ?", goalID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

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

		goalID := c.Param("id")
		if goalID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var goal models.Goal
		result = db.First(&goal, "id = ?", goalID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedGoal models.Goal
		if err := c.Bind(&updatedGoal); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedGoal.GoalTypeID != 0 {
			var goalType models.GoalType
			result = db.First(&goalType, "id = ?", updatedGoal.GoalTypeID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid goal type ID. Goal type not found."}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			updatedGoal.GoalTypeID = goalType.ID
			updatedGoal.GoalTypeName = goalType.GoalType

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

		db.Save(&goal)

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

		goalID := c.Param("id")
		if goalID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Goal ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var goal models.Goal
		result = db.First(&goal, "id = ?", goalID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&goal)

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

		var kpiIndicator models.KPIIndicator
		if err := c.Bind(&kpiIndicator); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if kpiIndicator.DesignationID == 0 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Designation ID is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var designation models.Designation
		result = db.First(&designation, kpiIndicator.DesignationID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Designation not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		kpiIndicator.DesignationName = designation.DesignationName

		if !helper.IsValidScore(kpiIndicator) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid score. Scores should be between 0 and 5"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		totalScores := helper.CalculateTotalScores(kpiIndicator)
		kpiIndicator.Result = totalScores / 37

		kpiIndicator.AdminId = adminUser.ID
		kpiIndicator.AdminName = adminUser.FirstName + " " + adminUser.LastName

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

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		searching := c.QueryParam("searching")

		query := db.Model(&models.KPIIndicator{})
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where(
				"LOWER(title) LIKE ? OR LOWER(designation_name) LIKE ? OR LOWER(admin_name) LIKE ?",
				searchPattern, searchPattern, searchPattern,
			)
		}

		var totalCount int64
		query.Count(&totalCount)

		var kpiIndicators []models.KPIIndicator
		if err := query.Offset(offset).Limit(perPage).Find(&kpiIndicators).Order("id DESC").Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching KPI indicators"})
		}

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

		performanceIDStr := c.Param("id")
		performanceID, err := strconv.ParseUint(performanceIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid KPI Indicator ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var kpiIndicator models.KPIIndicator
		result = db.First(&kpiIndicator, uint(performanceID))
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "KPI Indicator not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

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

		kpiIndicatorIDStr := c.Param("id")
		kpiIndicatorID, err := strconv.ParseUint(kpiIndicatorIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid KPI Indicator ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var kpiIndicator models.KPIIndicator
		result = db.First(&kpiIndicator, uint(kpiIndicatorID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "KPI Indicator not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updateData models.KPIIndicator
		if err := c.Bind(&updateData); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updateData.DesignationID != 0 {
			var newDesignation models.Designation
			result = db.First(&newDesignation, updateData.DesignationID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Designation not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			kpiIndicator.DesignationName = newDesignation.DesignationName
		}

		db.Model(&kpiIndicator).Updates(updateData)

		totalScores := helper.CalculateTotalScores(kpiIndicator)
		kpiIndicator.Result = totalScores / 37

		db.Save(&kpiIndicator)

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

		kpiIndicatorIDStr := c.Param("id")
		kpiIndicatorID, err := strconv.ParseUint(kpiIndicatorIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid KPI Indicator ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var kpiIndicator models.KPIIndicator
		result = db.First(&kpiIndicator, uint(kpiIndicatorID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "KPI Indicator not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&kpiIndicator)

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

		var kpaIndicator models.KPAIndicator
		if err := c.Bind(&kpaIndicator); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if kpaIndicator.EmployeeID == 0 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Employee ID is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var employee models.Employee
		result = db.First(&employee, kpaIndicator.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		kpaIndicator.EmployeeName = employee.FirstName + " " + employee.LastName

		if !helper.IsValidScoreKPA(kpaIndicator) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid score. Scores should be between 0 and 5"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		totalScores := helper.CalculateTotalScoresKPA(kpaIndicator)
		kpaIndicator.Result = totalScores / 37

		kpaIndicator.AdminId = adminUser.ID
		kpaIndicator.AdminName = adminUser.FirstName + " " + adminUser.LastName

		if kpaIndicator.AppraisalDate == "" || !helper.IsValidAppraisalDateFormat(kpaIndicator.AppraisalDate) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid appraisal date format. Please use mm-yyyy format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		appraisalTime, err := time.Parse("01-2006", kpaIndicator.AppraisalDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid appraisal date format. Please use mm-yyyy format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		kpaIndicator.AppraisalDate = appraisalTime.Format("2006-01")

		db.Create(&kpaIndicator)

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

func GetAllKPAIndicatorsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		searching := c.QueryParam("searching")

		query := db.Model(&models.KPAIndicator{})
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where(
				"LOWER(title) LIKE ? OR LOWER(employee_name) LIKE ? OR LOWER(admin_name) LIKE ? OR appraisal_date LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern,
			)
		}

		var totalCount int64
		query.Count(&totalCount)

		var kpaIndicators []models.KPAIndicator
		if err := query.Offset(offset).Limit(perPage).Find(&kpaIndicators).Order("id DESC").Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching KPA indicators"})
		}

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

		performanceIDStr := c.Param("id")
		performanceID, err := strconv.ParseUint(performanceIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid KPA Indicator ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var kpaIndicator models.KPAIndicator
		result = db.First(&kpaIndicator, uint(performanceID))
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "KPA Indicator not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

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

		kpaIndicatorIDStr := c.Param("id")
		kpaIndicatorID, err := strconv.ParseUint(kpaIndicatorIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid KPA Indicator ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var kpaIndicator models.KPAIndicator
		result = db.First(&kpaIndicator, uint(kpaIndicatorID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "KPA Indicator not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updateData models.KPAIndicator
		if err := c.Bind(&updateData); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updateData.EmployeeID != 0 {
			var newEmployee models.Employee
			result = db.First(&newEmployee, updateData.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			kpaIndicator.EmployeeName = newEmployee.FirstName + " " + newEmployee.LastName
		}

		db.Model(&kpaIndicator).Updates(updateData)

		totalScores := helper.CalculateTotalScoresKPA(kpaIndicator)
		kpaIndicator.Result = totalScores / 37

		db.Save(&kpaIndicator)

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

		kpaIndicatorIDStr := c.Param("id")
		kpaIndicatorID, err := strconv.ParseUint(kpaIndicatorIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid KPA Indicator ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var kpaIndicator models.KPAIndicator
		result = db.First(&kpaIndicator, uint(kpaIndicatorID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "KPA Indicator not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&kpaIndicator)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "KPA Indicator deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
