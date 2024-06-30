package controllers

import (
	"fmt"
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

func CreateOvertimeRequestByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var overtime models.OvertimeRequest
		if err := c.Bind(&overtime); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if overtime.Date == "" || overtime.InTime == "" || overtime.OutTime == "" || overtime.Reason == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid Overtime Request. All fields are required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(overtime.Reason) < 5 || len(overtime.Reason) > 3000 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Reason must be between 5 and 3000 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		overtime.EmployeeID = employee.ID
		overtime.Username = employee.Username
		overtime.FullNameEmployee = employee.FirstName + " " + employee.LastName
		overtime.Status = "Pending"

		_, err = time.Parse("2006-01-02", overtime.Date)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid Overtime Request date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

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
		totalWorkHours := workDuration.Hours()
		totalWorkMinutes := int(workDuration.Minutes())
		totalWork := strconv.FormatFloat(totalWorkHours, 'f', 2, 64) + " hours"
		overtime.TotalWork = totalWork
		overtime.TotalMinutes = totalWorkMinutes

		db.Create(&overtime)

		db.Preload("Employee").First(&overtime, overtime.ID)

		err = helper.SendOvertimeRequestNotification(employee.Email, overtime.FullNameEmployee, overtime.Date, overtime.InTime, overtime.OutTime, overtime.Reason)
		if err != nil {
			fmt.Println("Failed to send overtime request notification email:", err)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Overtime Request data added successfully",
			"data":    overtime,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

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

		// Retrieve employee details
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
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

		// Query parameters for searching
		searching := c.QueryParam("searching")

		// Build the query
		query := db.Model(&models.OvertimeRequest{}).Where("employee_id = ?", employee.ID)
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where(
				"LOWER(full_name_employee) LIKE ? OR LOWER(date) LIKE ? OR LOWER(in_time) LIKE ? OR LOWER(out_time) LIKE ? OR LOWER(status) LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
			)
		}

		// Retrieve overtime requests for the employee with pagination
		var overtimeRequests []models.OvertimeRequest
		result = query.Preload("Employee").Order("id DESC").Offset(offset).Limit(perPage).Find(&overtimeRequests)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch overtime requests"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Get total count of overtime requests for the employee
		var totalCount int64
		query.Count(&totalCount)

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Overtime requests retrieved successfully",
			"data":    overtimeRequests,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetOvertimeRequestByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		overtimeRequestIDStr := c.Param("id")
		overtimeRequestID, err := strconv.ParseUint(overtimeRequestIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid overtime request ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var overtimeRequest models.OvertimeRequest
		result = db.Preload("Employee").Where("id = ?", overtimeRequestID).First(&overtimeRequest)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Overtime request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if overtimeRequest.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Overtime request does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Overtime request retrieved successfully",
			"data":    overtimeRequest,
		})
	}
}

func UpdateOvertimeRequestByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		overtimeRequestIDStr := c.Param("id")
		overtimeRequestID, err := strconv.ParseUint(overtimeRequestIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid overtime request ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var overtimeRequest models.OvertimeRequest
		result = db.Where("id = ?", overtimeRequestID).First(&overtimeRequest)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Overtime request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if overtimeRequest.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Overtime request does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		var updatedOvertimeRequest models.OvertimeRequest
		if err := c.Bind(&updatedOvertimeRequest); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedOvertimeRequest.Date != "" {
			overtimeRequest.Date = updatedOvertimeRequest.Date
		}
		if updatedOvertimeRequest.InTime != "" {
			overtimeRequest.InTime = updatedOvertimeRequest.InTime
		}
		if updatedOvertimeRequest.OutTime != "" {
			overtimeRequest.OutTime = updatedOvertimeRequest.OutTime
		}

		/*
			if updatedOvertimeRequest.Reason != "" {
				overtimeRequest.Reason = updatedOvertimeRequest.Reason
			}
		*/

		if updatedOvertimeRequest.Reason != "" {
			if len(updatedOvertimeRequest.Reason) < 5 || len(updatedOvertimeRequest.Reason) > 3000 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Reason must be between 5 and 3000 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			overtimeRequest.Reason = updatedOvertimeRequest.Reason
		}

		if updatedOvertimeRequest.InTime != "" || updatedOvertimeRequest.OutTime != "" {
			inTime, _ := time.Parse("15:04", overtimeRequest.InTime)
			outTime, _ := time.Parse("15:04", overtimeRequest.OutTime)
			workDuration := outTime.Sub(inTime)
			totalWorkHours := workDuration.Hours()
			overtimeRequest.TotalWork = strconv.FormatFloat(totalWorkHours, 'f', 2, 64) + " hours"

			totalWorkMinutes := int(workDuration.Minutes())
			overtimeRequest.TotalMinutes = totalWorkMinutes
		}

		overtimeRequest.Status = overtimeRequest.Status

		db.Save(&overtimeRequest)

		db.Preload("Employee").First(&overtimeRequest, overtimeRequest.ID)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Overtime request updated successfully",
			"data":    overtimeRequest,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteOvertimeRequestByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		overtimeRequestIDStr := c.Param("id")
		overtimeRequestID, err := strconv.ParseUint(overtimeRequestIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid overtime request ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var overtimeRequest models.OvertimeRequest
		result = db.Where("id = ?", overtimeRequestID).First(&overtimeRequest)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Overtime request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if overtimeRequest.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Overtime request does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		db.Delete(&overtimeRequest)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Overtime request deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
