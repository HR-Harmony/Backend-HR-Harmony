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

func CreateLeaveRequestTypeByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Bind the leave request type data from the request body
		var leaveRequestType models.LeaveRequestType
		if err := c.Bind(&leaveRequestType); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate leave request type data
		if leaveRequestType.LeaveType == "" || leaveRequestType.DaysPerYears <= 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Incomplete leave request type data"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Create the leave request type in the database
		if err := db.Create(&leaveRequestType).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create leave request type"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Leave request type created successfully",
			"data":    leaveRequestType,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// GetAllLeaveRequestTypesByAdmin handles the retrieval of all leave request types by admin with pagination and searching
func GetAllLeaveRequestTypesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Fetch all leave request types from the database
		var leaveRequestTypes []models.LeaveRequestType
		db.Model(&models.LeaveRequestType{}).Offset(offset).Limit(perPage).Find(&leaveRequestTypes)

		// Check if searching query param is provided
		searching := c.QueryParam("searching")
		if searching != "" {
			var filteredLeaveRequestTypes []models.LeaveRequestType
			for _, lrt := range leaveRequestTypes {
				if strings.Contains(strings.ToLower(lrt.LeaveType), strings.ToLower(searching)) ||
					strings.Contains(strings.ToLower(fmt.Sprintf("%d", lrt.DaysPerYears)), strings.ToLower(searching)) ||
					strings.Contains(strings.ToLower(fmt.Sprintf("%t", lrt.IsRequiresApproval)), strings.ToLower(searching)) {
					filteredLeaveRequestTypes = append(filteredLeaveRequestTypes, lrt)
				}
			}
			leaveRequestTypes = filteredLeaveRequestTypes
		}

		var totalCount int64
		db.Model(&models.LeaveRequestType{}).Count(&totalCount)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":                http.StatusOK,
			"error":               false,
			"message":             "Leave request types retrieved successfully",
			"leave_request_types": leaveRequestTypes,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetLeaveRequestTypeByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve leave request type ID from the request URL parameter
		leaveRequestTypeID := c.Param("id")
		if leaveRequestTypeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Leave request type ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch the leave request type from the database based on the ID
		var leaveRequestType models.LeaveRequestType
		result = db.First(&leaveRequestType, "id = ?", leaveRequestTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave request type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":               http.StatusOK,
			"error":              false,
			"message":            "Leave request type retrieved successfully",
			"leave_request_type": leaveRequestType,
		})
	}
}

func UpdateLeaveRequestTypeByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve leave request type ID from the request URL parameter
		leaveRequestTypeID := c.Param("id")
		if leaveRequestTypeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Leave request type ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch the leave request type from the database based on the ID
		var leaveRequestType models.LeaveRequestType
		result = db.First(&leaveRequestType, "id = ?", leaveRequestTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave request type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated leave request type data from the request body
		var updatedLeaveRequestType models.LeaveRequestType
		if err := c.Bind(&updatedLeaveRequestType); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update the leave request type data
		if updatedLeaveRequestType.LeaveType != "" {
			leaveRequestType.LeaveType = updatedLeaveRequestType.LeaveType
		}
		if updatedLeaveRequestType.DaysPerYears != 0 {
			leaveRequestType.DaysPerYears = updatedLeaveRequestType.DaysPerYears
		}
		if updatedLeaveRequestType.IsRequiresApproval != leaveRequestType.IsRequiresApproval {
			leaveRequestType.IsRequiresApproval = updatedLeaveRequestType.IsRequiresApproval
		}

		// Update the leave request type in the database
		db.Save(&leaveRequestType)

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request type updated successfully",
			"data":    leaveRequestType,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteLeaveRequestTypeByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve leave request type ID from the request URL parameter
		leaveRequestTypeID := c.Param("id")
		if leaveRequestTypeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Leave request type ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch the leave request type from the database based on the ID
		var leaveRequestType models.LeaveRequestType
		result = db.First(&leaveRequestType, "id = ?", leaveRequestTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave request type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the leave request type from the database
		db.Delete(&leaveRequestType)

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request type deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func CreateLeaveRequestByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind the leave request data from the request body
		var leaveRequest models.LeaveRequest
		if err := c.Bind(&leaveRequest); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Add employee username based on employee ID
		var employee models.Employee
		result = db.First(&employee, "id = ?", leaveRequest.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		leaveRequest.Username = employee.Username
		leaveRequest.FullNameEmployee = employee.FirstName + " " + employee.LastName

		// Add leave type based on leave type ID
		var leaveType models.LeaveRequestType
		result = db.First(&leaveType, "id = ?", leaveRequest.LeaveTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		leaveRequest.LeaveType = leaveType.LeaveType

		// Parse start date from string to time.Time
		startDate, err := time.Parse("2006-01-02", leaveRequest.StartDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid StartDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Parse end date from string to time.Time
		endDate, err := time.Parse("2006-01-02", leaveRequest.EndDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid EndDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Calculate the difference in days
		days := int(endDate.Sub(startDate).Hours() / 24)

		// Assign the calculated days to leaveRequest
		leaveRequest.Days = days

		// Format start date in "yyyy-mm-dd" format
		leaveRequest.StartDate = startDate.Format("2006-01-02")

		// Format end date in "yyyy-mm-dd" format
		leaveRequest.EndDate = endDate.Format("2006-01-02")

		leaveRequest.Status = "Pending"

		// Save the leave request to the database
		if err := db.Create(&leaveRequest).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create leave request"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Leave request created successfully",
			"data":    leaveRequest,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// GetAllLeaveRequestsByAdmin handles the retrieval of all leave requests by admin with pagination and searching
func GetAllLeaveRequestsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Fetch searching query parameters
		searching := c.QueryParam("searching")

		// Fetch leave request data from database with optional search filters
		var leaveRequests []models.LeaveRequest
		query := db.Model(&leaveRequests)
		if searching != "" {
			query = query.Where("LOWER(username) LIKE ? OR LOWER(leave_type) LIKE ? OR LOWER(start_date) LIKE ? OR LOWER(end_date) LIKE ? OR days = ?",
				"%"+strings.ToLower(searching)+"%",
				"%"+strings.ToLower(searching)+"%",
				"%"+strings.ToLower(searching)+"%",
				"%"+strings.ToLower(searching)+"%",
				helper.ParseStringToInt(searching),
			)
		}
		query.Offset(offset).Limit(perPage).Find(&leaveRequests)

		var totalCount int64
		db.Model(&models.LeaveRequest{}).Count(&totalCount)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request data retrieved successfully",
			"data":    leaveRequests,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetLeaveRequestByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Retrieve leave request ID from the request URL parameter
		leaveRequestID := c.Param("id")
		if leaveRequestID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Leave request ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch the leave request from the database based on the ID
		var leaveRequest models.LeaveRequest
		result = db.First(&leaveRequest, "id = ?", leaveRequestID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request retrieved successfully",
			"data":    leaveRequest,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateLeaveRequestByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Retrieve leave request ID from the request URL parameter
		leaveRequestID := c.Param("id")
		if leaveRequestID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Leave request ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch the leave request from the database based on the ID
		var leaveRequest models.LeaveRequest
		result = db.First(&leaveRequest, "id = ?", leaveRequestID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated leave request data from the request body
		var updatedLeaveRequest models.LeaveRequest
		if err := c.Bind(&updatedLeaveRequest); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update the leave request data
		if updatedLeaveRequest.EmployeeID != 0 {
			// Update employee username based on the new employee ID
			var employee models.Employee
			result = db.First(&employee, "id = ?", updatedLeaveRequest.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			leaveRequest.Username = employee.Username
			leaveRequest.EmployeeID = updatedLeaveRequest.EmployeeID
			leaveRequest.FullNameEmployee = employee.FirstName + " " + employee.LastName
		}
		if updatedLeaveRequest.LeaveTypeID != 0 {
			// Update leave type based on the new leave type ID
			var leaveType models.LeaveRequestType
			result = db.First(&leaveType, "id = ?", updatedLeaveRequest.LeaveTypeID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave type not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			leaveRequest.LeaveType = leaveType.LeaveType
			leaveRequest.LeaveTypeID = updatedLeaveRequest.LeaveTypeID
		}
		if updatedLeaveRequest.StartDate != "" {
			// Parse start date from string to time.Time
			startDate, err := time.Parse("2006-01-02", updatedLeaveRequest.StartDate)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid StartDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			leaveRequest.StartDate = startDate.Format("2006-01-02")
		}
		if updatedLeaveRequest.EndDate != "" {
			// Parse end date from string to time.Time
			endDate, err := time.Parse("2006-01-02", updatedLeaveRequest.EndDate)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid EndDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			leaveRequest.EndDate = endDate.Format("2006-01-02")
		}
		if updatedLeaveRequest.IsHalfDay {
			leaveRequest.IsHalfDay = updatedLeaveRequest.IsHalfDay
		}
		if updatedLeaveRequest.Remarks != "" {
			leaveRequest.Remarks = updatedLeaveRequest.Remarks
		}
		if updatedLeaveRequest.LeaveReason != "" {
			leaveRequest.LeaveReason = updatedLeaveRequest.LeaveReason
		}

		if updatedLeaveRequest.Status != "" {
			leaveRequest.Status = updatedLeaveRequest.Status
		}

		// Update the days based on the updated start and end dates
		if leaveRequest.StartDate != "" && leaveRequest.EndDate != "" {
			startDate, _ := time.Parse("2006-01-02", leaveRequest.StartDate)
			endDate, _ := time.Parse("2006-01-02", leaveRequest.EndDate)
			days := int(endDate.Sub(startDate).Hours() / 24) // Calculate the difference in days
			leaveRequest.Days = days
		}

		// Save the updated leave request to the database
		if err := db.Save(&leaveRequest).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update leave request"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request updated successfully",
			"data":    leaveRequest,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteLeaveRequestByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Retrieve leave request ID from the request URL parameter
		leaveRequestID := c.Param("id")
		if leaveRequestID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Leave request ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch the leave request from the database based on the ID
		var leaveRequest models.LeaveRequest
		result = db.First(&leaveRequest, "id = ?", leaveRequestID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the leave request from the database
		if err := db.Delete(&leaveRequest).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete leave request"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
