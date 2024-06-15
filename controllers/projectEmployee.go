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

func AddProjectByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var project models.Project
		if err := c.Bind(&project); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var department models.Department
		if err := db.Where("id = ?", project.DepartmentID).First(&department).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid department ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var clientEmployee models.Employee
		if err := db.Where("id = ? AND is_client = ?", project.EmployeeID, true).First(&clientEmployee).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid client employee ID or employee is not a client"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		project.Username = clientEmployee.Username
		project.ClientName = clientEmployee.FirstName + " " + clientEmployee.LastName
		project.DepartmentName = department.DepartmentName

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

		if err := db.Create(&project).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create project"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Project added successfully",
			"project": project,
		})
	}
}

func GetAllProjectsByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var projects []models.Project
		if err := db.Find(&projects).Order("id DESC").Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch projects"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"message":  "List of projects retrieved successfully",
			"projects": projects,
		})
	}
}

func GetProjectByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid project ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var project models.Project
		if err := db.First(&project, projectID).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Project not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Project retrieved successfully",
			"project": project,
		})
	}
}

func UpdateProjectByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid project ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingProject models.Project
		result = db.First(&existingProject, projectID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Project not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedProject struct {
			Title         string `json:"title"`
			EmployeeID    uint   `json:"employee_id"`
			EstimatedHour int    `json:"estimated_hour"`
			Priority      string `json:"priority"`
			StartDate     string `json:"start_date"`
			EndDate       string `json:"end_date"`
			Summary       string `json:"summary"`
			DepartmentID  uint   `json:"department_id"`
			Description   string `json:"description"`
			Status        string `json:"status"`
			ProjectBar    *int   `json:"project_bar"`
		}

		if err := c.Bind(&updatedProject); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedProject.Title != "" {
			existingProject.Title = updatedProject.Title
		}

		if updatedProject.EmployeeID != 0 {
			var clientEmployee models.Employee
			result := db.Where("id = ? AND is_client = ?", updatedProject.EmployeeID, true).First(&clientEmployee)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid client employee ID or employee is not a client"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
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

		if updatedProject.ProjectBar != nil {
			existingProject.ProjectBar = *updatedProject.ProjectBar
		}

		currentTime := time.Now()
		existingProject.UpdatedAt = currentTime

		db.Save(&existingProject)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Project updated successfully",
			Project: &existingProject,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func UpdateProjectByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid project ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingProject models.Project
		result = db.First(&existingProject, projectID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Project not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedProject models.Project
		if err := c.Bind(&updatedProject); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedProject.Title != "" {
			existingProject.Title = updatedProject.Title
		}

		if updatedProject.EmployeeID != 0 {
			var clientEmployee models.Employee
			result := db.Where("id = ? AND is_client = ?", updatedProject.EmployeeID, true).First(&clientEmployee)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid client employee ID or employee is not a client"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
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

		currentTime := time.Now()
		existingProject.UpdatedAt = currentTime

		db.Save(&existingProject)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Project updated successfully",
			Project: &existingProject,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func DeleteProjectByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid project ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingProject models.Project
		result = db.First(&existingProject, projectID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Project not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&existingProject)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Project deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetProjectStatusByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Initialize the project status counts
		projectStatus := map[string]int{
			"Cancelled":   0,
			"Completed":   0,
			"Not_Started": 0,
			"On_Hold":     0,
			"In_Progress": 0,
		}

		// Query the project counts by status
		var projectStatusCounts []struct {
			Status string
			Count  int
		}
		if err := db.Model(&models.Project{}).Select("status, count(*) as count").Group("status").Scan(&projectStatusCounts).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve project counts by status"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Update the project status counts based on the query results
		for _, count := range projectStatusCounts {
			statusKey := strings.ReplaceAll(count.Status, " ", "_")
			projectStatus[statusKey] = count.Count
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":           http.StatusOK,
			"error":          false,
			"message":        "Project counts by status retrieved successfully",
			"project_status": projectStatus,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
