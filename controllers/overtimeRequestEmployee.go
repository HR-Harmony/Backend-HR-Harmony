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

// CreateOvertimeRequestForEmployee memungkinkan karyawan untuk menambahkan data overtime request untuk dirinya sendiri
func CreateOvertimeRequestByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Bind the overtime request data from the request body
		var overtime models.OvertimeRequest
		if err := c.Bind(&overtime); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate overtime request data
		if overtime.Date == "" || overtime.InTime == "" || overtime.OutTime == "" || overtime.Reason == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid Overtime Request. All fields are required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Set overtime request data
		overtime.EmployeeID = employee.ID
		overtime.Username = employee.Username
		overtime.Status = "Pending"

		// Validate date format
		_, err = time.Parse("2006-01-02", overtime.Date)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid Overtime Request date format. Required format: yyyy-mm-dd"}
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

		// Add total_work to overtime request data
		overtime.TotalWork = totalWork

		// Create the overtime request in the database
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

// GetOvertimeRequestsForEmployee mengambil semua data overtime request yang dimiliki oleh karyawan berdasarkan employee ID-nya
func GetAllOvertimeRequestsByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve overtime requests for the employee
		var overtimeRequests []models.OvertimeRequest
		db.Where("employee_id = ?", employee.ID).Find(&overtimeRequests)

		// Return the overtime requests
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Overtime requests retrieved successfully",
			"data":    overtimeRequests,
		})
	}
}

// GetOvertimeRequestByIDForEmployee mengambil data overtime request milik karyawan berdasarkan ID overtime request
func GetOvertimeRequestByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve overtime request ID from path parameter
		overtimeRequestIDStr := c.Param("id")
		overtimeRequestID, err := strconv.ParseUint(overtimeRequestIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid overtime request ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the overtime request
		var overtimeRequest models.OvertimeRequest
		result = db.Where("id = ?", overtimeRequestID).First(&overtimeRequest)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Overtime request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Check if the overtime request belongs to the employee
		if overtimeRequest.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Overtime request does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Return the overtime request
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Overtime request retrieved successfully",
			"data":    overtimeRequest,
		})
	}
}

// UpdateOvertimeRequestByIDForEmployee mengubah data overtime request milik karyawan berdasarkan ID overtime request
func UpdateOvertimeRequestByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve overtime request ID from path parameter
		overtimeRequestIDStr := c.Param("id")
		overtimeRequestID, err := strconv.ParseUint(overtimeRequestIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid overtime request ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the overtime request
		var overtimeRequest models.OvertimeRequest
		result = db.Where("id = ?", overtimeRequestID).First(&overtimeRequest)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Overtime request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Check if the overtime request belongs to the employee
		if overtimeRequest.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Overtime request does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind the updated overtime request data from the request body
		var updatedOvertimeRequest models.OvertimeRequest
		if err := c.Bind(&updatedOvertimeRequest); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update the overtime request data if provided
		if updatedOvertimeRequest.Date != "" {
			overtimeRequest.Date = updatedOvertimeRequest.Date
		}
		if updatedOvertimeRequest.InTime != "" {
			overtimeRequest.InTime = updatedOvertimeRequest.InTime
		}
		if updatedOvertimeRequest.OutTime != "" {
			overtimeRequest.OutTime = updatedOvertimeRequest.OutTime
		}
		if updatedOvertimeRequest.Reason != "" {
			overtimeRequest.Reason = updatedOvertimeRequest.Reason
		}

		// Calculate total work duration if in_time or out_time is updated
		if updatedOvertimeRequest.InTime != "" || updatedOvertimeRequest.OutTime != "" {
			inTime, _ := time.Parse("15:04", overtimeRequest.InTime)
			outTime, _ := time.Parse("15:04", overtimeRequest.OutTime)
			workDuration := outTime.Sub(inTime)
			totalWorkHours := workDuration.Hours()
			overtimeRequest.TotalWork = strconv.FormatFloat(totalWorkHours, 'f', 2, 64) + " hours"
		}

		// Save the updated overtime request data
		db.Save(&overtimeRequest)

		// Return success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Overtime request updated successfully",
			"data":    overtimeRequest,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// DeleteOvertimeRequestByIDForEmployee menghapus overtime request milik karyawan berdasarkan ID overtime request
func DeleteOvertimeRequestByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve overtime request ID from path parameter
		overtimeRequestIDStr := c.Param("id")
		overtimeRequestID, err := strconv.ParseUint(overtimeRequestIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid overtime request ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the overtime request
		var overtimeRequest models.OvertimeRequest
		result = db.Where("id = ?", overtimeRequestID).First(&overtimeRequest)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Overtime request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Check if the overtime request belongs to the employee
		if overtimeRequest.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Overtime request does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Delete the overtime request
		db.Delete(&overtimeRequest)

		// Return success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Overtime request deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
