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

		startDate, err := time.Parse("2006-01-02", task.StartDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid StartDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Parse end date from string to time.Time
		endDate, err := time.Parse("2006-01-02", task.EndDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid EndDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Format start date and end date in "yyyy-mm-dd" format
		task.StartDate = startDate.Format("2006-01-02")
		task.EndDate = endDate.Format("2006-01-02")

		// Check if the project with the given ID exists
		var existingProject models.Project
		result = db.First(&existingProject, task.ProjectID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Project not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Set the ProjectName
		task.ProjectName = existingProject.Title

		// Set the created timestamp
		currentTime := time.Now()
		task.CreatedAt = &currentTime

		// Create the task in the database
		db.Create(&task)

		// Respond with success
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

		// Retrieve all tasks from the database
		var tasks []models.Task
		db.Find(&tasks)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Tasks retrieved successfully",
			Tasks:   tasks,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetTaskByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Extract task ID from the request params
		taskIDStr := c.Param("id")
		taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid task ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve task from the database based on ID
		var task models.Task
		result = db.First(&task, uint(taskID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Task not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with success
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

		// Extract task ID from the request params
		taskIDStr := c.Param("id")
		taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid task ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve task from the database based on ID
		var existingTask models.Task
		result = db.First(&existingTask, uint(taskID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Task not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind updated task data from the request body
		var updatedTask models.Task
		if err := c.Bind(&updatedTask); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate that at least one field is updated
		if updatedTask.Title == "" && updatedTask.StartDate == "" && updatedTask.EndDate == "" && updatedTask.EstimatedHour == 0 &&
			updatedTask.ProjectID == 0 && updatedTask.Summary == "" && updatedTask.Description == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "At least one field must be updated"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update task data selectively
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
			// Update ProjectName based on the new ProjectID
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

		db.Save(&existingTask)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Task updated successfully",
			Task:    &existingTask,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteTaskByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Extract task ID from the request params
		taskIDStr := c.Param("id")
		taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid task ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve task from the database based on ID
		var existingTask models.Task
		result = db.First(&existingTask, uint(taskID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Task not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete task from the database
		db.Delete(&existingTask)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Task deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
