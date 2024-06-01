package controllers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type DashboardSummaryResponse struct {
	Code             int                      `json:"code"`
	Error            bool                     `json:"error"`
	Message          string                   `json:"message"`
	TotalOvertime    int64                    `json:"total_overtime"`
	TotalLeave       int64                    `json:"total_leave"`
	TotalProjects    int                      `json:"total_projects"`
	ProjectSummaries []map[string]interface{} `json:"project_summaries"`
	TotalTasks       int                      `json:"total_tasks"`
	TaskSummaries    []map[string]interface{} `json:"task_summaries"`
	PayrollSummary   []PayrollSummaryItem     `json:"payroll_summary"`
	TrainingSummary  []TrainingSummaryItem    `json:"training_summary"`
}

type DashboardSummaryItem struct {
	ID          uint    `json:"id"`
	Title       string  `json:"title"`
	ProgressBar float64 `json:"progress_bar"`
}

type TrainingSummaryItem struct {
	Month          string  `json:"month"`
	TotalTrainings float64 `json:"total_training"`
}

var (
	employeeCache *cache.Cache
)

// Initialize cache
func init() {
	employeeCache = cache.New(5*time.Minute, 10*time.Minute)
}

func GetDashboardSummaryForEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Check cache
		cacheKey := fmt.Sprintf("employeeDashboardSummary-%d", employee.ID)
		cached, found := employeeCache.Get(cacheKey)
		if found {
			return c.JSON(http.StatusOK, cached)
		}

		var totalOvertime int64
		db.Model(&models.OvertimeRequest{}).Where("employee_id = ?", employee.ID).Count(&totalOvertime)

		var totalLeave int64
		db.Model(&models.LeaveRequest{}).Where("employee_id = ?", employee.ID).Count(&totalLeave)

		var projects []models.Project
		if err := db.Find(&projects).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve project data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		projectSummaries := make([]map[string]interface{}, 0)
		for _, project := range projects {
			projectSummary := map[string]interface{}{
				"project_id":   project.ID,
				"project_name": project.Title,
				"project_bar":  project.ProjectBar,
			}
			projectSummaries = append(projectSummaries, projectSummary)
		}

		var tasks []models.Task
		if err := db.Find(&tasks).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to retrieve task data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		taskSummaries := make([]map[string]interface{}, 0)
		for _, task := range tasks {
			taskSummary := map[string]interface{}{
				"task_id":      task.ID,
				"title":        task.Title,
				"progress_bar": task.ProgressBar,
			}
			taskSummaries = append(taskSummaries, taskSummary)
		}

		currentYear := time.Now().Year()
		payrollSummary := make([]PayrollSummaryItem, 0, 12)

		for month := time.January; month <= time.December; month++ {
			startOfMonth := time.Date(currentYear, month, 1, 0, 0, 0, 0, time.UTC)
			endOfMonth := startOfMonth.AddDate(0, 1, -1).Add(24 * time.Hour)

			var totalBasicSalary float64
			if err := db.Model(&models.PayrollInfo{}).
				Where("employee_id = ? AND created_at BETWEEN ? AND ?", employee.ID, startOfMonth, endOfMonth).
				Select("COALESCE(SUM(basic_salary), 0)").
				Row().
				Scan(&totalBasicSalary); err != nil {
				return err
			}

			payrollSummary = append(payrollSummary, PayrollSummaryItem{
				Month:  startOfMonth.Format("January 2006"),
				Amount: totalBasicSalary,
			})
		}

		trainingSummary := make([]TrainingSummaryItem, 0, 12)

		for month := time.January; month <= time.December; month++ {
			startOfMonth := fmt.Sprintf("%d-%02d-01", currentYear, month)
			endOfMonth := time.Date(currentYear, month+1, 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

			var totalTrainings int64
			if err := db.Model(&models.Training{}).
				Where("employee_id = ? AND start_date BETWEEN ? AND ?", employee.ID, startOfMonth, endOfMonth).
				Count(&totalTrainings).Error; err != nil {
				return err
			}

			trainingSummary = append(trainingSummary, TrainingSummaryItem{
				Month:          time.Month(month).String() + " " + strconv.Itoa(currentYear),
				TotalTrainings: float64(totalTrainings),
			})
		}

		response := DashboardSummaryResponse{
			Code:             http.StatusOK,
			Error:            false,
			Message:          "Dashboard summary retrieved successfully",
			TotalOvertime:    totalOvertime,
			TotalLeave:       totalLeave,
			TotalProjects:    len(projectSummaries),
			ProjectSummaries: projectSummaries,
			TotalTasks:       len(taskSummaries),
			TaskSummaries:    taskSummaries,
			PayrollSummary:   payrollSummary,
			TrainingSummary:  trainingSummary,
		}

		employeeCache.Set(cacheKey, response, cache.DefaultExpiration)

		return c.JSON(http.StatusOK, response)
	}
}

/*

// GetTotalOvertimeRequestsForEmployee mengambil total jumlah overtime request yang dimiliki oleh seorang karyawan berdasarkan employee ID-nya

func GetTotalOvertimeRequestsForEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Count the total number of overtime requests for the employee
		var totalCount int64
		db.Model(&models.OvertimeRequest{}).Where("employee_id = ?", employee.ID).Count(&totalCount)

		// Return the total count
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Total overtime requests retrieved successfully",
			"data":    totalCount,
		})
	}
}

// GetTotalLeaveRequestsForEmployee mengambil total jumlah leave request yang dimiliki oleh seorang karyawan berdasarkan employee ID-nya
func GetTotalLeaveRequestsForEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Count the total number of leave requests for the employee
		var totalCount int64
		db.Model(&models.LeaveRequest{}).Where("employee_id = ?", employee.ID).Count(&totalCount)

		// Return the total count
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Total leave requests retrieved successfully",
			"data":    totalCount,
		})
	}
}

// Project Grafik
func GetProjectSummaryForEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
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

// Task Grafik
func GetTaskSummaryForEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
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

// GetPayrollSummaryForEmployee retrieves payroll summary for employee dashboard
func GetPayrollSummaryForEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
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
				Where("employee_id = ? AND created_at BETWEEN ? AND ?", employee.ID, startOfMonth, endOfMonth).
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

// TrainingSummaryItem represents an item in training summary
type TrainingSummaryItem struct {
	Month          string  `json:"month"`
	TotalTrainings float64 `json:"total_training"`
}

// GetTrainingSummaryForEmployee retrieves training summary for employee dashboard
func GetTrainingSummaryForEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Retrieve training summary for the current year
		currentYear := time.Now().Year()
		trainingSummary := make([]TrainingSummaryItem, 0, 12)

		for month := time.January; month <= time.December; month++ {
			// Get the start and end of the month
			startOfMonth := fmt.Sprintf("%d-%02d-01", currentYear, month)
			endOfMonth := time.Date(currentYear, month+1, 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

			// Retrieve total number of trainings for the month
			var totalTrainings int64
			if err := db.Model(&models.Training{}).
				Where("employee_id = ? AND start_date BETWEEN ? AND ?", employee.ID, startOfMonth, endOfMonth).
				Count(&totalTrainings).Error; err != nil {
				return err
			}

			// Store the total number of trainings for the month
			trainingSummary = append(trainingSummary, TrainingSummaryItem{
				Month:          time.Month(month).String() + " " + strconv.Itoa(currentYear),
				TotalTrainings: float64(totalTrainings),
			})
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":             http.StatusOK,
			"error":            false,
			"message":          "Training summary retrieved successfully",
			"training_summary": trainingSummary,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

*/
