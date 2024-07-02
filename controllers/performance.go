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
		if err := query.Order("id DESC").Offset(offset).Limit(perPage).Find(&goalTypes).Error; err != nil {
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

		// Check if the goal type is associated with any goals
		var goalCount int64
		db.Model(&models.Goal{}).Where("goal_type_id = ?", goalTypeID).Count(&goalCount)
		if goalCount > 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Cannot delete goal type because it is associated with one or more goals"}
			return c.JSON(http.StatusBadRequest, errorResponse)
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

type GoalResponse struct {
	ID                uint   `json:"id"`
	GoalTypeID        uint   `json:"goal_type_id"`
	GoalTypeName      string `json:"goal_type_name"`
	ProjectID         uint   `json:"project_id"`
	ProjectName       string `json:"project_name"`
	TaskID            uint   `json:"task_id"`
	TaskName          string `json:"task_name"`
	TrainingID        uint   `json:"training_id"`
	TrainingSkillID   uint   `json:"training_skill_id"`
	TrainingSkillName string `json:"training_skill_name"`
	Subject           string `json:"subject"`
	TargetAchievement string `json:"target_achievement"`
	StartDate         string `json:"start_date"`
	EndDate           string `json:"end_date"`
	Description       string `json:"description"`
	GoalRating        uint   `json:"goal_rating"`
	ProgressBar       uint   `json:"progress_bar"`
	Status            string `json:"status"`
	CreatedAt         string `json:"created_at"`
}

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

		response := GoalResponse{
			ID:                goal.ID,
			GoalTypeID:        goal.GoalTypeID,
			GoalTypeName:      goal.GoalTypeName,
			ProjectID:         goal.ProjectID,
			ProjectName:       goal.ProjectName,
			TaskID:            goal.TaskID,
			TaskName:          goal.TaskName,
			TrainingID:        goal.TrainingID,
			TrainingSkillID:   goal.TrainingSkillID,
			TrainingSkillName: goal.TrainingSkillName,
			Subject:           goal.Subject,
			TargetAchievement: goal.TargetAchievement,
			StartDate:         goal.StartDate,
			EndDate:           goal.EndDate,
			Description:       goal.Description,
			GoalRating:        goal.GoalRating,
			ProgressBar:       goal.ProgressBar,
			Status:            goal.Status,
			CreatedAt:         goal.CreatedAt,
		}

		successResponse := map[string]interface{}{
			"Code":    http.StatusOK,
			"Message": "Goal added successfully",
			"Data":    response,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
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
*/

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
		if err := query.Order("id DESC").Offset(offset).Limit(perPage).Find(&goals).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching goals"})
		}

		var goalsResponse []GoalResponse
		for _, goal := range goals {
			goalsResponse = append(goalsResponse, GoalResponse{
				ID:                goal.ID,
				GoalTypeID:        goal.GoalTypeID,
				GoalTypeName:      goal.GoalTypeName,
				ProjectID:         goal.ProjectID,
				ProjectName:       goal.ProjectName,
				TaskID:            goal.TaskID,
				TaskName:          goal.TaskName,
				TrainingID:        goal.TrainingID,
				TrainingSkillID:   goal.TrainingSkillID,
				TrainingSkillName: goal.TrainingSkillName,
				Subject:           goal.Subject,
				TargetAchievement: goal.TargetAchievement,
				StartDate:         goal.StartDate,
				EndDate:           goal.EndDate,
				Description:       goal.Description,
				GoalRating:        goal.GoalRating,
				ProgressBar:       goal.ProgressBar,
				Status:            goal.Status,
				CreatedAt:         goal.CreatedAt,
			})
		}

		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "All goals retrieved successfully",
			"goals":      goalsResponse,
			"pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
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
		if err := query.Order("id DESC").Offset(offset).Limit(perPage).Find(&goals).Error; err != nil {
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
*/

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
		result = db.First(&goal, goalID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Goal not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		goalResponse := GoalResponse{
			ID:                goal.ID,
			GoalTypeID:        goal.GoalTypeID,
			ProjectID:         goal.ProjectID,
			TaskID:            goal.TaskID,
			TrainingID:        goal.TrainingID,
			TrainingSkillID:   goal.TrainingSkillID,
			Subject:           goal.Subject,
			TargetAchievement: goal.TargetAchievement,
			StartDate:         goal.StartDate,
			EndDate:           goal.EndDate,
			Description:       goal.Description,
			GoalRating:        goal.GoalRating,
			ProgressBar:       goal.ProgressBar,
			Status:            goal.Status,
			CreatedAt:         goal.CreatedAt,
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Goal retrieved successfully",
			"data":    goalResponse,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
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
*/

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
			updatedGoal.GoalTypeName = goalType.GoalType
		}

		if updatedGoal.ProjectID != 0 {
			var project models.Project
			result = db.First(&project, updatedGoal.ProjectID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Project not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			updatedGoal.ProjectName = project.Title
		}

		if updatedGoal.TaskID != 0 {
			var task models.Task
			result = db.First(&task, updatedGoal.TaskID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Task not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			updatedGoal.TaskName = task.Title
		}

		if updatedGoal.TrainingID != 0 {
			var training models.Training
			result = db.First(&training, updatedGoal.TrainingID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Training not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
		}

		if updatedGoal.TrainingSkillID != 0 {
			var trainingSkill models.TrainingSkill
			result = db.First(&trainingSkill, updatedGoal.TrainingSkillID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Training Skill ID not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			updatedGoal.TrainingSkillName = trainingSkill.TrainingSkill
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

		// Prepare response
		response := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Goal updated successfully",
			"data": GoalResponse{
				ID:                goal.ID,
				GoalTypeID:        goal.GoalTypeID,
				GoalTypeName:      goal.GoalTypeName,
				ProjectID:         goal.ProjectID,
				ProjectName:       goal.ProjectName,
				TaskID:            goal.TaskID,
				TaskName:          goal.TaskName,
				TrainingID:        goal.TrainingID,
				TrainingSkillID:   goal.TrainingSkillID,
				TrainingSkillName: goal.TrainingSkillName,
				Subject:           goal.Subject,
				TargetAchievement: goal.TargetAchievement,
				StartDate:         goal.StartDate,
				EndDate:           goal.EndDate,
				Description:       goal.Description,
				GoalRating:        goal.GoalRating,
				ProgressBar:       goal.ProgressBar,
				Status:            goal.Status,
				CreatedAt:         goal.CreatedAt,
			},
		}
		return c.JSON(http.StatusOK, response)
	}
}

/*
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
*/

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

// GetAllKPIIndicatorsByAdmin handles the retrieval of all KPI indicators by admin with pagination
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
		if err := query.Order("id DESC").Offset(offset).Limit(perPage).Find(&kpiIndicators).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching KPI indicators"})
		}

		// Batch processing for designation_name
		designationMap := make(map[uint]string)
		designationIDs := make([]uint, 0, len(kpiIndicators))

		for _, kpi := range kpiIndicators {
			if _, found := designationMap[kpi.DesignationID]; !found {
				designationIDs = append(designationIDs, kpi.DesignationID)
			}
		}

		var designations []models.Designation
		db.Model(&models.Designation{}).Where("id IN (?)", designationIDs).Find(&designations)

		// Create map for fast lookup
		for _, des := range designations {
			designationMap[des.ID] = des.DesignationName
		}

		// Update designation_name field
		tx := db.Begin()
		for i := range kpiIndicators {
			kpiIndicators[i].DesignationName = designationMap[kpiIndicators[i].DesignationID]

			if err := tx.Save(&kpiIndicators[i]).Error; err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error saving KPI indicators"})
			}
		}

		tx.Commit()

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

/*
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
		if err := query.Order("id DESC").Offset(offset).Limit(perPage).Find(&kpiIndicators).Error; err != nil {
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
*/

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

// GetAllKPAIndicatorsByAdmin handles the retrieval of all KPA indicators by admin with pagination
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
		if err := query.Order("id DESC").Offset(offset).Limit(perPage).Find(&kpaIndicators).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching KPA indicators"})
		}

		// Batch processing for employee_name
		employeeMap := make(map[uint]string)
		employeeIDs := make([]uint, 0, len(kpaIndicators))

		for _, kpa := range kpaIndicators {
			if _, found := employeeMap[kpa.EmployeeID]; !found {
				employeeIDs = append(employeeIDs, kpa.EmployeeID)
			}
		}

		var employees []models.Employee
		db.Model(&models.Employee{}).Where("id IN (?)", employeeIDs).Find(&employees)

		// Create map for fast lookup
		for _, emp := range employees {
			employeeMap[emp.ID] = emp.FullName // Menggunakan FullName, bukan FullNameEmployee
		}

		// Update employee_name field
		tx := db.Begin()
		for i := range kpaIndicators {
			kpaIndicators[i].EmployeeName = employeeMap[kpaIndicators[i].EmployeeID]

			if err := tx.Save(&kpaIndicators[i]).Error; err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error saving KPA indicators"})
			}
		}

		tx.Commit()

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

/*
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
		if err := query.Order("id DESC").Offset(offset).Limit(perPage).Find(&kpaIndicators).Error; err != nil {
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
*/

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
