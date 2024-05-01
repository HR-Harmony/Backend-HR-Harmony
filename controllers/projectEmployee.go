// controllers/addProject.go

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

// AddProject handles the addition of a new project by an employee
func AddProjectByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Check if the user is an employee
		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Parse request body to get project details
		var project models.Project
		if err := c.Bind(&project); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate department ID
		var department models.Department
		if err := db.Where("id = ?", project.DepartmentID).First(&department).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid department ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the provided employee ID exists and is a client
		var clientEmployee models.Employee
		if err := db.Where("id = ? AND is_client = ?", project.EmployeeID, true).First(&clientEmployee).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid client employee ID or employee is not a client"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Set username, client name, and department name
		project.Username = clientEmployee.Username
		project.ClientName = clientEmployee.FirstName + " " + clientEmployee.LastName
		project.DepartmentName = department.DepartmentName

		// Parse start date and end date strings to time.Time format
		startDate, err := time.Parse("2006-01-02", project.StartDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid start date format. Use yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		endDate, err := time.Parse("2006-01-02", project.EndDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid end date format. Use yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		project.StartDate = startDate.Format("2006-01-02")
		project.EndDate = endDate.Format("2006-01-02")

		project.Status = "Not Started"

		// Save the project to the database
		if err := db.Create(&project).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create project"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Return success response
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Project added successfully",
			"project": project,
		})
	}
}

// ListProjects handles the retrieval of all projects
func GetAllProjectsByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Check if the user is an employee
		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Retrieve all projects from the database
		var projects []models.Project
		if err := db.Find(&projects).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch projects"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Return the projects
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"message":  "List of projects retrieved successfully",
			"projects": projects,
		})
	}
}

// GetProjectByID handles the retrieval of a project by its ID
func GetProjectByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Check if the user is an employee
		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Get project ID from URL parameter
		projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid project ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the project from the database
		var project models.Project
		if err := db.First(&project, projectID).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Project not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Return the project
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Project retrieved successfully",
			"project": project,
		})
	}
}

// UpdateProjectByIDByEmployee handles the update of a project by an employee based on its ID
func UpdateProjectByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Check if the user is an employee
		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Get project ID from the URL parameters
		projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid project ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the project by ID
		var existingProject models.Project
		result = db.First(&existingProject, projectID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Project not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated project data from the request body
		var updatedProject models.Project
		if err := c.Bind(&updatedProject); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update only the specified fields (if provided)
		if updatedProject.Title != "" {
			existingProject.Title = updatedProject.Title
		}

		if updatedProject.EmployeeID != 0 {
			// Check if the provided employee ID is a client
			var clientEmployee models.Employee
			result := db.Where("id = ? AND is_client = ?", updatedProject.EmployeeID, true).First(&clientEmployee)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid client employee ID or employee is not a client"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			// Update the project with the new employee details
			existingProject.EmployeeID = updatedProject.EmployeeID
			existingProject.Username = clientEmployee.Username
			existingProject.ClientName = clientEmployee.FirstName + " " + clientEmployee.LastName
		}

		if updatedProject.EstimatedHour != 0 {
			existingProject.EstimatedHour = updatedProject.EstimatedHour
		}

		if updatedProject.Priority != "" {
			existingProject.Priority = updatedProject.Priority
		}

		if updatedProject.StartDate != "" {
			existingProject.StartDate = updatedProject.StartDate
		}

		if updatedProject.EndDate != "" {
			existingProject.EndDate = updatedProject.EndDate
		}

		if updatedProject.Summary != "" {
			existingProject.Summary = updatedProject.Summary
		}

		if updatedProject.DepartmentID != 0 {
			existingProject.DepartmentID = updatedProject.DepartmentID

			// Update department_name based on the new department_id
			var department models.Department
			result := db.First(&department, updatedProject.DepartmentID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			existingProject.DepartmentName = department.DepartmentName
		}

		if updatedProject.Description != "" {
			existingProject.Description = updatedProject.Description
		}

		if updatedProject.Status != "" {
			existingProject.Status = updatedProject.Status
		}

		if updatedProject.ProjectBar != 0 {
			existingProject.ProjectBar = updatedProject.ProjectBar
		}

		// Set the updated timestamp
		currentTime := time.Now()
		existingProject.UpdatedAt = currentTime

		// Save the changes to the database
		db.Save(&existingProject)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Project updated successfully",
			Project: &existingProject,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// DeleteProjectByIDByEmployee handles the deletion of a project by an employee based on its ID
func DeleteProjectByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Check if the user is an employee
		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Get project ID from the URL parameters
		projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid project ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the project by ID
		var existingProject models.Project
		result = db.First(&existingProject, projectID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Project not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the project from the database
		db.Delete(&existingProject)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Project deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
