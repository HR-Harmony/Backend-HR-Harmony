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

// CreateProjectByAdmin handles the creation of a new project by admin
func CreateProjectByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Bind the project data from the request body
		var project models.Project
		if err := c.Bind(&project); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate project data
		if project.Title == "" || project.Summary == "" || project.Description == "" || project.StartDate == "" || project.EndDate == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "All fields are required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Parse start date from string to time.Time
		startDate, err := time.Parse("2006-01-02", project.StartDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid StartDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Parse end date from string to time.Time
		endDate, err := time.Parse("2006-01-02", project.EndDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid EndDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Format start date and end date in "yyyy-mm-dd" format
		project.StartDate = startDate.Format("2006-01-02")
		project.EndDate = endDate.Format("2006-01-02")

		// Check if the employee with the given ID exists
		var existingEmployee models.Employee
		result = db.First(&existingEmployee, project.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Check if the department with the given ID exists
		var existingDepartment models.Department
		result = db.First(&existingDepartment, project.DepartmentID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Set the Username and DepartmentName
		project.Username = existingEmployee.Username
		project.DepartmentName = existingDepartment.DepartmentName
		project.ClientName = existingEmployee.FirstName + " " + existingEmployee.LastName

		project.Status = "Not Started"

		// Set the created timestamp
		currentTime := time.Now()
		project.CreatedAt = &currentTime

		// Create the project in the database
		db.Create(&project)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusCreated,
			Error:   false,
			Message: "Project created successfully",
			Project: &project,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// GetAllProjectsByAdmin handles the retrieval of all projects by admin with pagination
func GetAllProjectsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve projects from the database with pagination
		var projects []models.Project
		db.Offset(offset).Limit(perPage).Find(&projects)

		// Get total count of projects
		var totalCount int64
		db.Model(&models.Project{}).Count(&totalCount)

		// Respond with success
		successResponse := map[string]interface{}{
			"Code":       http.StatusOK,
			"Error":      false,
			"Message":    "Projects retrieved successfully",
			"Projects":   projects,
			"Pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetProjectByIDByAdmin handles the retrieval of a project by admin based on its ID
func GetProjectByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get project ID from the URL parameters
		projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid project ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the project by ID
		var project models.Project
		result = db.First(&project, projectID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Project not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Project retrieved successfully",
			Project: &project,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// UpdateProjectByIDByAdmin handles the update of a project by admin based on its ID
func UpdateProjectByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			existingProject.EmployeeID = updatedProject.EmployeeID

			// Update username based on the new employee_id
			var employee models.Employee
			result := db.First(&employee, updatedProject.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			existingProject.Username = employee.Username
			existingProject.ClientName = employee.FirstName + " " + employee.LastName
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

// DeleteProjectByIDByAdmin handles the deletion of a project by admin based on its ID
func DeleteProjectByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
