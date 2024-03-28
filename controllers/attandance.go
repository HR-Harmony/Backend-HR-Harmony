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

func AddManualAttendanceByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Bind the attendance data from the request body
		var attendance models.Attendance
		if err := c.Bind(&attendance); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate attendance data
		if attendance.EmployeeID == 0 || attendance.AttendanceDate == "" || attendance.InTime == "" || attendance.OutTime == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance data. All fields are required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the employee ID exists
		var employee models.Employee
		result = db.First(&employee, attendance.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee ID not found"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		attendance.Username = employee.Username

		// Validate attandance_date format
		_, err = time.Parse("2006-01-02", attendance.AttendanceDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Calculate total work duration
		inTime, err := time.Parse("15:04", attendance.InTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid in_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		outTime, err := time.Parse("15:04", attendance.OutTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid out_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		workDuration := outTime.Sub(inTime)

		// Convert work duration to hours
		totalWorkHours := workDuration.Hours()

		// Convert totalWorkHours to string
		totalWork := strconv.FormatFloat(totalWorkHours, 'f', 2, 64) + " hours"

		// Add total_work to attendance data
		attendance.TotalWork = totalWork

		// Create the attendance in the database
		db.Create(&attendance)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Attendance data added successfully",
			"data":    attendance,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// GetAttendanceByAdmin handles the retrieval of attendance data by admin
func GetAllAttendanceByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get query parameters
		date := c.QueryParam("date")
		employeeID := c.QueryParam("employee_id")

		// Prepare query based on filters
		query := db.Model(&models.Attendance{})

		if date != "" {
			query = query.Where("attendance_date = ?", date)
		}

		if employeeID != "" {
			query = query.Where("employee_id = ?", employeeID)
		}

		// Fetch attendance data
		var attendance []models.Attendance
		result = query.Find(&attendance)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch attendance data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Attendance data retrieved successfully",
			"data":    attendance,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetAttendanceByIDByAdmin handles the retrieval of attendance data by admin based on attendance ID
func GetAttendanceByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get attendance ID from path parameter
		attendanceID := c.Param("id")
		if attendanceID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Attendance ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch attendance data by ID
		var attendance models.Attendance
		result = db.First(&attendance, "id = ?", attendanceID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Attendance not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Attendance data retrieved successfully",
			"data":    attendance,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateAttendanceByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get attendance ID from path parameter
		attendanceID := c.Param("id")
		if attendanceID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Attendance ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch attendance data by ID
		var attendance models.Attendance
		result = db.First(&attendance, "id = ?", attendanceID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Attendance not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind updated attendance data from the request body
		var updatedAttendance models.Attendance
		if err := c.Bind(&updatedAttendance); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update fields that are allowed to be changed
		if updatedAttendance.EmployeeID != 0 {
			// Fetch employee data by ID
			var employee models.Employee
			result = db.First(&employee, "id = ?", updatedAttendance.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid employee ID. Employee not found."}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			attendance.EmployeeID = updatedAttendance.EmployeeID
			attendance.Username = employee.Username
		}
		if updatedAttendance.AttendanceDate != "" {
			// Validate attendance_date format
			_, err := time.Parse("2006-01-02", updatedAttendance.AttendanceDate)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance date format. Required format: yyyy-mm-dd"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			attendance.AttendanceDate = updatedAttendance.AttendanceDate
		}
		if updatedAttendance.InTime != "" {
			attendance.InTime = updatedAttendance.InTime
		}
		if updatedAttendance.OutTime != "" {
			attendance.OutTime = updatedAttendance.OutTime
		}

		// Recalculate total work duration
		inTime, err := time.Parse("15:04", attendance.InTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid in_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		outTime, err := time.Parse("15:04", attendance.OutTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid out_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		workDuration := outTime.Sub(inTime)

		// Convert work duration to hours
		totalWorkHours := workDuration.Hours()

		// Convert totalWorkHours to string
		totalWork := strconv.FormatFloat(totalWorkHours, 'f', 2, 64) + " hours"

		// Update total work
		attendance.TotalWork = totalWork

		// Save the updated attendance data to the database
		db.Save(&attendance)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Attendance data updated successfully",
			"data":    attendance,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteAttendanceByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get attendance ID from path parameter
		attendanceID := c.Param("id")
		if attendanceID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Attendance ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch attendance data by ID
		var attendance models.Attendance
		result = db.First(&attendance, "id = ?", attendanceID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Attendance not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the attendance data from the database
		db.Delete(&attendance)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Attendance data deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func CreateOvertimeRequestByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Bind the attendance data from the request body
		var overtime models.OvertimeRequest
		if err := c.Bind(&overtime); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate attendance data
		if overtime.EmployeeID == 0 || overtime.Date == "" || overtime.InTime == "" || overtime.OutTime == "" || overtime.Reason == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid Overtime Request. All fields are required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the employee ID exists
		var employee models.Employee
		result = db.First(&employee, overtime.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee ID not found"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		overtime.Username = employee.Username

		// Validate attandance_date format
		_, err = time.Parse("2006-01-02", overtime.Date)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid Overtime Request date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate attandance_date format
		_, err = time.Parse("2006-01-02", overtime.Date)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Calculate total work duration
		inTime, err := time.Parse("15:04", overtime.InTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid in_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		outTime, err := time.Parse("15:04", overtime.OutTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid out_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		workDuration := outTime.Sub(inTime)

		// Convert work duration to hours
		totalWorkHours := workDuration.Hours()

		// Convert totalWorkHours to string
		totalWork := strconv.FormatFloat(totalWorkHours, 'f', 2, 64) + " hours"

		// Add total_work to attendance data
		overtime.TotalWork = totalWork

		overtime.Status = "Pending"

		// Create the attendance in the database
		db.Create(&overtime)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Overtime Request data added successfully",
			"data":    overtime,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllOvertimeRequestsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get query parameters
		date := c.QueryParam("date")
		employeeID := c.QueryParam("employee_id")

		// Prepare query based on filters
		query := db.Model(&models.OvertimeRequest{})

		if date != "" {
			query = query.Where("date = ?", date)
		}

		if employeeID != "" {
			query = query.Where("employee_id = ?", employeeID)
		}

		// Fetch attendance data
		var overtime []models.OvertimeRequest
		result = query.Find(&overtime)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch Overtime Request data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Overtime Request  retrieved successfully",
			"data":    overtime,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetOvertimeRequestByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get attendance ID from path parameter
		overtimeID := c.Param("id")
		if overtimeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Overtime ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch attendance data by ID
		var overtime models.OvertimeRequest
		result = db.First(&overtime, "id = ?", overtimeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Overtime Request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Attendance data retrieved successfully",
			"data":    overtime,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateOvertimeRequestByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get attendance ID from path parameter
		overtimeID := c.Param("id")
		if overtimeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Overtime ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch attendance data by ID
		var overtime models.OvertimeRequest
		result = db.First(&overtime, "id = ?", overtimeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Overtime Request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind updated attendance data from the request body
		var updatedOvertime models.OvertimeRequest
		if err := c.Bind(&updatedOvertime); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update fields that are allowed to be changed
		if updatedOvertime.EmployeeID != 0 {
			// Fetch employee data by ID
			var employee models.Employee
			result = db.First(&employee, "id = ?", updatedOvertime.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid employee ID. Employee not found."}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			overtime.EmployeeID = updatedOvertime.EmployeeID
			overtime.Username = employee.Username
		}
		if updatedOvertime.Date != "" {
			// Validate attendance_date format
			_, err := time.Parse("2006-01-02", updatedOvertime.Date)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid Overtime Request date format. Required format: yyyy-mm-dd"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			overtime.Date = updatedOvertime.Date
		}
		if updatedOvertime.InTime != "" {
			overtime.InTime = updatedOvertime.InTime
		}
		if updatedOvertime.OutTime != "" {
			overtime.OutTime = updatedOvertime.OutTime
		}

		if updatedOvertime.Reason != "" {
			overtime.Reason = updatedOvertime.Reason
		}

		if updatedOvertime.Status != "" {
			overtime.Status = updatedOvertime.Status
		}

		// Validate attandance_date format
		_, err = time.Parse("2006-01-02", overtime.Date)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Calculate total work duration
		inTime, err := time.Parse("15:04", overtime.InTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid in_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		outTime, err := time.Parse("15:04", overtime.OutTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid out_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		workDuration := outTime.Sub(inTime)

		// Convert work duration to hours
		totalWorkHours := workDuration.Hours()

		// Convert totalWorkHours to string
		totalWork := strconv.FormatFloat(totalWorkHours, 'f', 2, 64) + " hours"

		// Add total_work to attendance data
		overtime.TotalWork = totalWork

		// Save the updated attendance data to the database
		db.Save(&overtime)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Attendance data updated successfully",
			"data":    overtime,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteOvertimeRequestByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get attendance ID from path parameter
		overtimeID := c.Param("id")
		if overtimeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Overtime ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch attendance data by ID
		var overtime models.OvertimeRequest
		result = db.First(&overtime, "id = ?", overtimeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Overtime Request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the attendance data from the database
		db.Delete(&overtime)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Overtime Request data deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
