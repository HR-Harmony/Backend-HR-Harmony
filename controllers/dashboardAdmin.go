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

// DashboardSummary represents the aggregated data for admin dashboard
type DashboardSummary struct {
	ProjectStatus     map[string]int           `json:"project_status"`
	Departments       map[string]int           `json:"departments"`
	Designations      map[string]int           `json:"designations"`
	AttendanceSummary map[string]int           `json:"attendance_summary"`
	ProjectSummary    []map[string]interface{} `json:"project_summary"`
	TaskSummary       []map[string]interface{} `json:"task_summary"`
	PayrollSummary    []PayrollSummaryItem     `json:"payroll_summary"`
}

// PayrollSummaryItem represents an item in payroll summary
type PayrollSummaryItem struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}

// GetDashboardSummaryForAdmin retrieves aggregated dashboard summary for admin dashboard
func GetDashboardSummaryForAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve project counts by status
		var projectStatusCounts []struct {
			Status string
			Count  int
		}
		if err := db.Model(&models.Project{}).Select("status, count(*) as count").Group("status").Scan(&projectStatusCounts).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve project counts by status"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Map project status counts to a more readable format
		projectStatus := make(map[string]int)
		for _, count := range projectStatusCounts {
			projectStatus[count.Status] = count.Count
		}

		// Retrieve all departments
		var departments []models.Department
		if err := db.Find(&departments).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve departments"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Retrieve employee count for each department
		departmentEmployeeCounts := make(map[string]int)
		for _, department := range departments {
			var employeeCount int64
			if err := db.Model(&models.Employee{}).Where("department_id = ?", department.ID).Count(&employeeCount).Error; err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve employee count for department"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
			departmentEmployeeCounts[department.DepartmentName] = int(employeeCount)
		}

		// Retrieve all designations
		var designations []models.Designation
		if err := db.Find(&designations).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve designations"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Retrieve employee count for each designation
		designationEmployeeCounts := make(map[string]int)
		for _, designation := range designations {
			var employeeCount int64
			if err := db.Model(&models.Employee{}).Where("designation_id = ?", designation.ID).Count(&employeeCount).Error; err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve employee count for designation"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
			designationEmployeeCounts[designation.DesignationName] = int(employeeCount)
		}

		// Get current date
		currentDate := time.Now().Format("2006-01-02")

		// Fetch attendance data for today for non-client employees
		var presentCount int64
		if err := db.Model(&models.Attendance{}).
			Joins("JOIN employees ON attendances.employee_id = employees.id").
			Where("attendances.attendance_date = ? AND employees.is_client = ?", currentDate, false).
			Count(&presentCount).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch attendance data for today"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Fetch total number of non-client employees
		var totalStaffCount int64
		if err := db.Model(&models.Employee{}).Where("is_client = ?", false).Count(&totalStaffCount).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch total staff count"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Calculate absent count
		absentCount := totalStaffCount - presentCount

		// Retrieve project summary
		var projects []models.Project
		if err := db.Find(&projects).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve project data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Map project data to project summaries
		projectSummaries := make([]map[string]interface{}, 0)
		for _, project := range projects {
			projectSummary := map[string]interface{}{
				"project_id":   project.ID,
				"project_name": project.Title,
				"project_bar":  project.ProjectBar,
			}
			projectSummaries = append(projectSummaries, projectSummary)
		}

		// Retrieve task summary
		var tasks []models.Task
		if err := db.Find(&tasks).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve task data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Map task data to task summaries
		taskSummaries := make([]map[string]interface{}, 0)
		for _, task := range tasks {
			taskSummary := map[string]interface{}{
				"task_id":      task.ID,
				"title":        task.Title,
				"progress_bar": task.ProgressBar,
			}
			taskSummaries = append(taskSummaries, taskSummary)
		}

		// Retrieve payroll summary for the current year
		currentYear := time.Now().Year()
		payrollSummary := make([]PayrollSummaryItem, 0, 12)

		for month := time.January; month <= time.December; month++ {
			// Get the start and end of the month
			startOfMonth := time.Date(currentYear, month, 1, 0, 0, 0, 0, time.UTC)
			endOfMonth := startOfMonth.AddDate(0, 1, -1).Add(24 * time.Hour)

			// Retrieve total basic salary for the month
			var totalBasicSalary float64
			if err := db.Model(&models.PayrollInfo{}).
				Where("created_at BETWEEN ? AND ?", startOfMonth, endOfMonth).
				Select("COALESCE(SUM(basic_salary), 0)").
				Row().
				Scan(&totalBasicSalary); err != nil {
				return err
			}

			// Store the total basic salary for the month
			payrollSummary = append(payrollSummary, PayrollSummaryItem{
				Month:  startOfMonth.Format("January 2006"),
				Amount: totalBasicSalary,
			})
		}

		// Construct aggregated dashboard summary
		dashboardSummary := DashboardSummary{
			ProjectStatus:     projectStatus,
			Departments:       departmentEmployeeCounts,
			Designations:      designationEmployeeCounts,
			AttendanceSummary: map[string]int{"total_staff": int(totalStaffCount), "present": int(presentCount), "absent": int(absentCount)},
			ProjectSummary:    projectSummaries,
			TaskSummary:       taskSummaries,
			PayrollSummary:    payrollSummary,
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":      http.StatusOK,
			"error":     false,
			"message":   "Dashboard summary retrieved successfully",
			"dashboard": dashboardSummary,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// Project Total With Status

// GetProjectStatusByAdmin handles the retrieval of project counts by status for admin dashboard

/*
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

		// Retrieve project counts by status
		var counts []struct {
			Status string
			Count  int
		}
		if err := db.Model(&models.Project{}).Select("status, count(*) as count").Group("status").Scan(&counts).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve project counts by status"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Map counts to a more readable format
		countMap := make(map[string]int)
		for _, count := range counts {
			countMap[count.Status] = count.Count
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":           http.StatusOK,
			"error":          false,
			"message":        "Project counts by status retrieved successfully",
			"project_counts": countMap,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetAllDepartmentsWithEmployeeCount retrieves all departments with employee count
func GetAllDepartmentsWithEmployeeCount(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve all departments
		var departments []models.Department
		if err := db.Find(&departments).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve departments"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Retrieve employee count for each department
		departmentEmployeeCounts := make(map[string]int)
		for _, department := range departments {
			var employeeCount int64
			if err := db.Model(&models.Employee{}).Where("department_id = ?", department.ID).Count(&employeeCount).Error; err != nil {
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve employee count for department"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
			departmentEmployeeCounts[department.DepartmentName] = int(employeeCount)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":        http.StatusOK,
			"error":       false,
			"message":     "Departments with employee count retrieved successfully",
			"departments": departmentEmployeeCounts,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetAllDesignationsWithEmployeeCount retrieves all designations with employee count
func GetAllDesignationsWithEmployeeCount(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve all designations
		var designations []models.Designation
		if err := db.Find(&designations).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve designations"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Retrieve employee count for each designation
		designationEmployeeCounts := make(map[string]int)
		for _, designation := range designations {
			var employeeCount int64
			if err := db.Model(&models.Employee{}).Where("designation_id = ?", designation.ID).Count(&employeeCount).Error; err != nil {
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve employee count for designation"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
			designationEmployeeCounts[designation.DesignationName] = int(employeeCount)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":         http.StatusOK,
			"error":        false,
			"message":      "Designations with employee count retrieved successfully",
			"designations": designationEmployeeCounts,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetAttendanceSummaryForToday retrieves attendance summary for today
func GetAttendanceSummaryForToday(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get current date
		currentDate := time.Now().Format("2006-01-02")

		// Fetch attendance data for today for non-client employees
		var presentCount int64
		if err := db.Model(&models.Attendance{}).
			Joins("JOIN employees ON attendances.employee_id = employees.id").
			Where("attendances.attendance_date = ? AND employees.is_client = ?", currentDate, false).
			Count(&presentCount).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch attendance data for today"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Fetch total number of non-client employees
		var totalStaffCount int64
		if err := db.Model(&models.Employee{}).Where("is_client = ?", false).Count(&totalStaffCount).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch total staff count"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Calculate absent count
		absentCount := totalStaffCount - presentCount

		// Respond with success
		successResponse := map[string]interface{}{
			"code":        http.StatusOK,
			"error":       false,
			"message":     "Attendance summary for today retrieved successfully",
			"total_staff": totalStaffCount,
			"present":     presentCount,
			"absent":      absentCount,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetProjectSummaryForAdmin retrieves project summary for admin dashboard
func GetProjectSummaryForAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve project summary
		var projects []models.Project
		if err := db.Find(&projects).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve project data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Map project data to project summaries
		projectSummaries := make([]map[string]interface{}, 0)
		for _, project := range projects {
			projectSummary := map[string]interface{}{
				"project_id":   project.ID,
				"project_name": project.Title,
				"project_bar":  project.ProjectBar,
			}
			projectSummaries = append(projectSummaries, projectSummary)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":           http.StatusOK,
			"error":          false,
			"message":        "Project summary retrieved successfully",
			"total_projects": len(projectSummaries),
			"projects":       projectSummaries,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetTaskSummaryForAdmin retrieves task summary for admin dashboard
func GetTaskSummaryForAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve task summary
		var tasks []models.Task
		if err := db.Find(&tasks).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve task data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Map task data to task summaries
		taskSummaries := make([]map[string]interface{}, 0)
		for _, task := range tasks {
			taskSummary := map[string]interface{}{
				"task_id":      task.ID,
				"title":        task.Title,
				"progress_bar": task.ProgressBar,
			}
			taskSummaries = append(taskSummaries, taskSummary)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":           http.StatusOK,
			"error":          false,
			"message":        "Task summary retrieved successfully",
			"total_tasks":    len(taskSummaries),
			"task_summaries": taskSummaries,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// PayrollSummaryItem represents an item in payroll summary
type PayrollSummaryItem struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}

// GetPayrollSummaryForAdmin retrieves payroll summary for admin dashboard
func GetPayrollSummaryForAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve payroll summary for the current year
		currentYear := time.Now().Year()
		payrollSummary := make([]PayrollSummaryItem, 0, 12)

		for month := time.January; month <= time.December; month++ {
			// Get the start and end of the month
			startOfMonth := time.Date(currentYear, month, 1, 0, 0, 0, 0, time.UTC)
			endOfMonth := startOfMonth.AddDate(0, 1, -1).Add(24 * time.Hour)

			// Retrieve total basic salary for the month
			var totalBasicSalary float64
			if err := db.Model(&models.PayrollInfo{}).
				Where("created_at BETWEEN ? AND ?", startOfMonth, endOfMonth).
				Select("COALESCE(SUM(basic_salary), 0)").
				Row().
				Scan(&totalBasicSalary); err != nil {
				return err
			}

			// Store the total basic salary for the month
			payrollSummary = append(payrollSummary, PayrollSummaryItem{
				Month:  startOfMonth.Format("January 2006"),
				Amount: totalBasicSalary,
			})
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":            http.StatusOK,
			"error":           false,
			"message":         "Payroll summary retrieved successfully",
			"payroll_summary": payrollSummary,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/
