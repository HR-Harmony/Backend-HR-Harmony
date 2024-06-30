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

type ProjectResponse struct {
	ID             uint       `json:"id"`
	Title          string     `json:"title"`
	EmployeeID     uint       `json:"employee_id"`
	Username       string     `json:"username"`
	ClientName     string     `json:"client_name"`
	EstimatedHour  int        `json:"estimated_hour"`
	Priority       string     `json:"priority"`
	StartDate      string     `json:"start_date"`
	EndDate        string     `json:"end_date"`
	Summary        string     `json:"summary"`
	DepartmentID   uint       `json:"department_id"`
	DepartmentName string     `json:"department_name"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	ProjectBar     int        `json:"project_bar"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func CreateProjectByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var project models.Project
		if err := c.Bind(&project); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if project.Title == "" || project.Summary == "" || project.Description == "" || project.StartDate == "" || project.EndDate == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "All fields are required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(project.Title) < 5 || len(project.Title) > 100 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Title must be between 5 and 100 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(project.Summary) < 5 || len(project.Summary) > 300 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Summary must be between 5 and 300 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(project.Description) < 5 || len(project.Description) > 3000 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Description must be between 5 and 3000 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		startDate, err := time.Parse("2006-01-02", project.StartDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid StartDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		endDate, err := time.Parse("2006-01-02", project.EndDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid EndDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		project.StartDate = startDate.Format("2006-01-02")
		project.EndDate = endDate.Format("2006-01-02")

		var existingEmployee models.Employee
		result = db.First(&existingEmployee, project.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var existingDepartment models.Department
		result = db.First(&existingDepartment, project.DepartmentID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		project.Username = existingEmployee.Username
		project.DepartmentName = existingDepartment.DepartmentName
		project.ClientName = existingEmployee.FirstName + " " + existingEmployee.LastName

		project.Status = "Not Started"

		currentTime := time.Now()
		project.CreatedAt = &currentTime

		db.Create(&project)

		// Map data from Project to ProjectResponse
		projectResponse := ProjectResponse{
			ID:             project.ID,
			Title:          project.Title,
			EmployeeID:     project.EmployeeID,
			Username:       project.Username,
			ClientName:     project.ClientName,
			EstimatedHour:  project.EstimatedHour,
			Priority:       project.Priority,
			StartDate:      project.StartDate,
			EndDate:        project.EndDate,
			Summary:        project.Summary,
			DepartmentID:   project.DepartmentID,
			DepartmentName: project.DepartmentName,
			Description:    project.Description,
			Status:         project.Status,
			ProjectBar:     project.ProjectBar,
			CreatedAt:      project.CreatedAt,
			UpdatedAt:      project.UpdatedAt,
		}

		successResponse := map[string]interface{}{
			"Code":    http.StatusCreated,
			"Error":   false,
			"Message": "Project created successfully",
			"Project": &projectResponse,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

/*
func CreateProjectByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var project models.Project
		if err := c.Bind(&project); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if project.Title == "" || project.Summary == "" || project.Description == "" || project.StartDate == "" || project.EndDate == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "All fields are required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(project.Title) < 5 || len(project.Title) > 100 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Title must be between 5 and 100 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(project.Summary) < 5 || len(project.Summary) > 300 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Summary must be between 5 and 300 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(project.Description) < 5 || len(project.Description) > 3000 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Description must be between 5 and 3000 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		startDate, err := time.Parse("2006-01-02", project.StartDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid StartDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		endDate, err := time.Parse("2006-01-02", project.EndDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid EndDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		project.StartDate = startDate.Format("2006-01-02")
		project.EndDate = endDate.Format("2006-01-02")

		var existingEmployee models.Employee
		result = db.First(&existingEmployee, project.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var existingDepartment models.Department
		result = db.First(&existingDepartment, project.DepartmentID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		project.Username = existingEmployee.Username
		project.DepartmentName = existingDepartment.DepartmentName
		project.ClientName = existingEmployee.FirstName + " " + existingEmployee.LastName

		project.Status = "Not Started"

		currentTime := time.Now()
		project.CreatedAt = &currentTime

		db.Create(&project)

		successResponse := helper.Response{
			Code:    http.StatusCreated,
			Error:   false,
			Message: "Project created successfully",
			Project: &project,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}
*/

// GetAllProjectsByAdmin retrieves all projects for an admin user
func GetAllProjectsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var projects []models.Project
		db.Order("id DESC").Offset(offset).Limit(perPage).Find(&projects)

		var totalCount int64
		db.Model(&models.Project{}).Count(&totalCount)

		// Map data from Project to ProjectResponse
		var projectResponses []ProjectResponse
		for _, project := range projects {
			projectResponse := ProjectResponse{
				ID:             project.ID,
				Title:          project.Title,
				EmployeeID:     project.EmployeeID,
				Username:       project.Username,
				ClientName:     project.ClientName,
				EstimatedHour:  project.EstimatedHour,
				Priority:       project.Priority,
				StartDate:      project.StartDate,
				EndDate:        project.EndDate,
				Summary:        project.Summary,
				DepartmentID:   project.DepartmentID,
				DepartmentName: project.DepartmentName,
				Description:    project.Description,
				Status:         project.Status,
				ProjectBar:     project.ProjectBar,
				CreatedAt:      project.CreatedAt,
				UpdatedAt:      project.UpdatedAt,
			}
			projectResponses = append(projectResponses, projectResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"Code":       http.StatusOK,
			"Error":      false,
			"Message":    "Projects retrieved successfully",
			"Projects":   projectResponses,
			"Pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func GetAllProjectsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var projects []models.Project
		db.Order("id DESC").Offset(offset).Limit(perPage).Find(&projects)

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
*/

func GetProjectByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid project ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var project models.Project
		result = db.First(&project, projectID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Project not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Map data from Project to ProjectResponse
		projectResponse := ProjectResponse{
			ID:             project.ID,
			Title:          project.Title,
			EmployeeID:     project.EmployeeID,
			Username:       project.Username,
			ClientName:     project.ClientName,
			EstimatedHour:  project.EstimatedHour,
			Priority:       project.Priority,
			StartDate:      project.StartDate,
			EndDate:        project.EndDate,
			Summary:        project.Summary,
			DepartmentID:   project.DepartmentID,
			DepartmentName: project.DepartmentName,
			Description:    project.Description,
			Status:         project.Status,
			ProjectBar:     project.ProjectBar,
			CreatedAt:      project.CreatedAt,
			UpdatedAt:      project.UpdatedAt,
		}

		successResponse := map[string]interface{}{
			"Code":    http.StatusOK,
			"Error":   false,
			"Message": "Project retrieved successfully",
			"Project": &projectResponse,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func GetProjectByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid project ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var project models.Project
		result = db.First(&project, projectID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Project not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Project retrieved successfully",
			Project: &project,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func UpdateProjectByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			if len(updatedProject.Title) < 5 || len(updatedProject.Title) > 100 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Title must be between 5 and 100"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingProject.Title = updatedProject.Title
		}

		if updatedProject.EmployeeID != 0 {
			existingProject.EmployeeID = updatedProject.EmployeeID

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
			if len(updatedProject.Summary) < 5 || len(updatedProject.Summary) > 300 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Summary must be between 5 and 300"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
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
			if len(updatedProject.Description) < 5 || len(updatedProject.Description) > 3000 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Description must be between 5 and 3000"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
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

		// Map data from Project to ProjectResponse
		projectResponse := ProjectResponse{
			ID:             existingProject.ID,
			Title:          existingProject.Title,
			EmployeeID:     existingProject.EmployeeID,
			Username:       existingProject.Username,
			ClientName:     existingProject.ClientName,
			EstimatedHour:  existingProject.EstimatedHour,
			Priority:       existingProject.Priority,
			StartDate:      existingProject.StartDate,
			EndDate:        existingProject.EndDate,
			Summary:        existingProject.Summary,
			DepartmentID:   existingProject.DepartmentID,
			DepartmentName: existingProject.DepartmentName,
			Description:    existingProject.Description,
			Status:         existingProject.Status,
			ProjectBar:     existingProject.ProjectBar,
			CreatedAt:      existingProject.CreatedAt,
			UpdatedAt:      existingProject.UpdatedAt,
		}

		successResponse := map[string]interface{}{
			"Code":    http.StatusOK,
			"Error":   false,
			"Message": "Project updated successfully",
			"Project": &projectResponse,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func UpdateProjectByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			if len(updatedProject.Title) < 5 || len(updatedProject.Title) > 100 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Title must be between 5 and 100"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingProject.Title = updatedProject.Title
		}

		if updatedProject.EmployeeID != 0 {
			existingProject.EmployeeID = updatedProject.EmployeeID

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
			if len(updatedProject.Summary) < 5 || len(updatedProject.Summary) > 300 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Summary must be between 5 and 300"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
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
			if len(updatedProject.Description) < 5 || len(updatedProject.Description) > 3000 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Description must be between 5 and 3000"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
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
*/

/*
func UpdateProjectByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			existingProject.EmployeeID = updatedProject.EmployeeID

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

func DeleteProjectByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

func GetProjectStatusByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
