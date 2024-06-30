package controllers

import (
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

func CreateNoteByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var note models.Note
		if err := c.Bind(&note); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		note.Fullname = adminUser.FirstName + " " + adminUser.LastName

		var existingTask models.Task
		result = db.First(&existingTask, note.TaskID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Task not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		currentTime := time.Now()
		note.CreatedAt = &currentTime

		db.Create(&note)

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Note created successfully",
			"note":    &note,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func DeleteNoteForTaskByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		noteID := c.Param("id")

		var note models.Note
		result = db.First(&note, noteID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Note not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&note)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Note deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func CreateTaskByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var task models.Task
		if err := c.Bind(&task); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if task.Title == "" || task.Summary == "" || task.Description == "" || task.StartDate == "" || task.EndDate == "" || task.ProjectID == 0 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "All fields are required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(task.Title) < 5 || len(task.Title) > 100 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Title must be between 5 and 100 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(task.Summary) < 5 || len(task.Summary) > 300 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Summary must be between 5 and 300 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(task.Description) < 5 || len(task.Description) > 3000 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "description must be between 5 and 3000 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		startDate, err := time.Parse("2006-01-02", task.StartDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid StartDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		endDate, err := time.Parse("2006-01-02", task.EndDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid EndDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		task.StartDate = startDate.Format("2006-01-02")
		task.EndDate = endDate.Format("2006-01-02")

		var existingProject models.Project
		result = db.First(&existingProject, task.ProjectID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Project not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		task.ProjectName = existingProject.Title
		task.Status = "Not Started"

		currentTime := time.Now()
		task.CreatedAt = &currentTime

		db.Create(&task)

		db.Preload("Project.Employee").First(&task, task.ID)

		successResponse := helper.Response{
			Code:    http.StatusCreated,
			Error:   false,
			Message: "Task created successfully",
			Task:    &task,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllTasksByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		var tasks []models.Task
		result = db.Preload("Notes").Preload("Project.Employee").Order("id DESC").Offset(offset).Limit(perPage).Find(&tasks)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve tasks"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var totalCount int64
		db.Model(&models.Task{}).Count(&totalCount)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Tasks retrieved successfully",
			"tasks":   tasks,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetTaskByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		taskIDStr := c.Param("id")
		taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid task ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var task models.Task
		result = db.Preload("Notes").Preload("Project.Employee").First(&task, uint(taskID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Task not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Task retrieved successfully",
			Task:    &task,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateTaskByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		taskIDStr := c.Param("id")
		taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid task ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingTask models.Task
		result = db.Preload("Notes").First(&existingTask, uint(taskID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Task not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedTask struct {
			Title         string `json:"title"`
			StartDate     string `json:"start_date"`
			EndDate       string `json:"end_date"`
			EstimatedHour int    `json:"estimated_hour"`
			ProjectID     uint   `json:"project_id"`
			Summary       string `json:"summary"`
			Description   string `json:"description"`
			Status        string `json:"status"`
			ProgressBar   *int   `json:"progress_bar"`
		}

		if err := c.Bind(&updatedTask); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedTask.Title == "" && updatedTask.StartDate == "" && updatedTask.EndDate == "" && updatedTask.EstimatedHour == 0 &&
			updatedTask.ProjectID == 0 && updatedTask.Summary == "" && updatedTask.Description == "" && updatedTask.ProgressBar == nil && updatedTask.Status == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "At least one field must be updated"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		/*
			if updatedTask.Title != "" {
				existingTask.Title = updatedTask.Title
			}
		*/

		if updatedTask.Title != "" {
			if len(updatedTask.Title) < 5 || len(updatedTask.Title) > 100 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Title must be between 5 and 100"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingTask.Title = updatedTask.Title
		}

		if updatedTask.StartDate != "" {
			startDate, err := time.Parse("2006-01-02", updatedTask.StartDate)
			if err != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid StartDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingTask.StartDate = startDate.Format("2006-01-02")
		}
		if updatedTask.EndDate != "" {
			endDate, err := time.Parse("2006-01-02", updatedTask.EndDate)
			if err != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid EndDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingTask.EndDate = endDate.Format("2006-01-02")
		}
		if updatedTask.EstimatedHour != 0 {
			existingTask.EstimatedHour = updatedTask.EstimatedHour
		}
		if updatedTask.ProjectID != 0 {
			existingTask.ProjectID = updatedTask.ProjectID
			var existingProject models.Project
			result := db.First(&existingProject, updatedTask.ProjectID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Project not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			existingTask.ProjectName = existingProject.Title
		}

		/*
			if updatedTask.Summary != "" {
				existingTask.Summary = updatedTask.Summary
			}
		*/

		if updatedTask.Summary != "" {
			if len(updatedTask.Summary) < 5 || len(updatedTask.Summary) > 300 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Summary must be between 5 and 300"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingTask.Summary = updatedTask.Summary
		}

		/*
			if updatedTask.Description != "" {
				existingTask.Description = updatedTask.Description
			}
		*/

		if updatedTask.Description != "" {
			if len(updatedTask.Description) < 5 || len(updatedTask.Description) > 3000 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Description must be between 5 and 3000"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingTask.Description = updatedTask.Description
		}

		if updatedTask.Status != "" {
			existingTask.Status = updatedTask.Status
		}

		if updatedTask.ProgressBar != nil {
			existingTask.ProgressBar = *updatedTask.ProgressBar
		}

		db.Save(&existingTask)

		db.Preload("Notes").Preload("Project.Employee").First(&existingTask, existingTask.ID)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Task updated successfully",
			Task:    &existingTask,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func UpdateTaskByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		taskIDStr := c.Param("id")
		taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid task ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingTask models.Task
		result = db.Preload("Notes").First(&existingTask, uint(taskID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Task not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedTask models.Task
		if err := c.Bind(&updatedTask); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedTask.Title == "" && updatedTask.StartDate == "" && updatedTask.EndDate == "" && updatedTask.EstimatedHour == 0 &&
			updatedTask.ProjectID == 0 && updatedTask.Summary == "" && updatedTask.Description == "" && updatedTask.ProgressBar == 0 && updatedTask.Status == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "At least one field must be updated"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedTask.Title != "" {
			existingTask.Title = updatedTask.Title
		}
		if updatedTask.StartDate != "" {
			startDate, err := time.Parse("2006-01-02", updatedTask.StartDate)
			if err != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid StartDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingTask.StartDate = startDate.Format("2006-01-02")
		}
		if updatedTask.EndDate != "" {
			endDate, err := time.Parse("2006-01-02", updatedTask.EndDate)
			if err != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid EndDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingTask.EndDate = endDate.Format("2006-01-02")
		}
		if updatedTask.EstimatedHour != 0 {
			existingTask.EstimatedHour = updatedTask.EstimatedHour
		}
		if updatedTask.ProjectID != 0 {
			existingTask.ProjectID = updatedTask.ProjectID
			var existingProject models.Project
			result := db.First(&existingProject, updatedTask.ProjectID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Project not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			existingTask.ProjectName = existingProject.Title
		}
		if updatedTask.Summary != "" {
			existingTask.Summary = updatedTask.Summary
		}
		if updatedTask.Description != "" {
			existingTask.Description = updatedTask.Description
		}

		if updatedTask.Status != "" {
			existingTask.Status = updatedTask.Status
		}

		if updatedTask.ProgressBar != 0 {
			existingTask.ProgressBar = updatedTask.ProgressBar
		}

		db.Save(&existingTask)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Task updated successfully",
			Task:    &existingTask,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func DeleteTaskByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		taskIDStr := c.Param("id")
		taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid task ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingTask models.Task
		result = db.Preload("Notes").First(&existingTask, uint(taskID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Task not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		for _, note := range existingTask.Notes {
			db.Delete(&note)
		}

		db.Delete(&existingTask)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Task deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetTaskStatusByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Initialize the task status counts
		taskStatus := map[string]int{
			"Cancelled":   0,
			"Completed":   0,
			"Not_Started": 0,
			"On_Hold":     0,
			"In_Progress": 0,
		}

		// Query the task counts by status
		var taskStatusCounts []struct {
			Status string
			Count  int
		}
		if err := db.Model(&models.Task{}).Select("status, count(*) as count").Group("status").Scan(&taskStatusCounts).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve task counts by status"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Update the task status counts based on the query results
		for _, count := range taskStatusCounts {
			statusKey := strings.ReplaceAll(count.Status, " ", "_")
			taskStatus[statusKey] = count.Count
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":        http.StatusOK,
			"error":       false,
			"message":     "Task counts by status retrieved successfully",
			"task_status": taskStatus,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
