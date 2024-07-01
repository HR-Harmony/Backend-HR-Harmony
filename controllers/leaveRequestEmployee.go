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

func CreateLeaveRequestByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind request body to leave request model
		var leaveRequest models.LeaveRequest
		if err := c.Bind(&leaveRequest); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Set employee details to leave request
		leaveRequest.EmployeeID = employee.ID
		leaveRequest.Username = employee.Username
		leaveRequest.FullNameEmployee = employee.FirstName + " " + employee.LastName

		// Retrieve leave type details
		var leaveType models.LeaveRequestType
		result = db.First(&leaveType, "id = ?", leaveRequest.LeaveTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		leaveRequest.LeaveType = leaveType.LeaveType

		// Calculate and set days
		startDate, err := time.Parse("2006-01-02", leaveRequest.StartDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid StartDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		endDate, err := time.Parse("2006-01-02", leaveRequest.EndDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid EndDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if leaveRequest.IsHalfDay {
			endDate = startDate
		}

		days := endDate.Sub(startDate).Hours() / 24

		if leaveRequest.IsHalfDay {
			days = 0.5
		}

		leaveRequest.Days = days
		leaveRequest.StartDate = startDate.Format("2006-01-02")
		leaveRequest.EndDate = endDate.Format("2006-01-02")
		leaveRequest.Status = "Pending"

		// Create leave request in the database
		if err := db.Create(&leaveRequest).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create leave request"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Send notification email to employee
		err = helper.SendLeaveRequestNotification(employee.Email, leaveRequest.FullNameEmployee, leaveRequest.LeaveType, leaveRequest.StartDate, leaveRequest.EndDate, leaveRequest.Days)
		if err != nil {
			fmt.Println("Failed to send leave request notification:", err)
		}

		// Prepare response in LeaveRequestResponse format
		response := LeaveRequestResponse{
			ID:               leaveRequest.ID,
			EmployeeID:       leaveRequest.EmployeeID,
			Username:         leaveRequest.Username,
			FullNameEmployee: leaveRequest.FullNameEmployee,
			LeaveTypeID:      leaveRequest.LeaveTypeID,
			LeaveType:        leaveRequest.LeaveType,
			StartDate:        leaveRequest.StartDate,
			EndDate:          leaveRequest.EndDate,
			IsHalfDay:        leaveRequest.IsHalfDay,
			Remarks:          leaveRequest.Remarks,
			LeaveReason:      leaveRequest.LeaveReason,
			Days:             leaveRequest.Days,
			Status:           leaveRequest.Status,
			CreatedAt:        leaveRequest.CreatedAt,
			UpdatedAt:        leaveRequest.UpdatedAt,
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Leave request created successfully",
			"data":    response,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

/*
func CreateLeaveRequestByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var leaveRequest models.LeaveRequest
		if err := c.Bind(&leaveRequest); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		leaveRequest.EmployeeID = employee.ID
		leaveRequest.Username = employee.Username
		leaveRequest.FullNameEmployee = employee.FirstName + " " + employee.LastName

		var leaveType models.LeaveRequestType
		result = db.First(&leaveType, "id = ?", leaveRequest.LeaveTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		leaveRequest.LeaveType = leaveType.LeaveType

		startDate, err := time.Parse("2006-01-02", leaveRequest.StartDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid StartDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		endDate, err := time.Parse("2006-01-02", leaveRequest.EndDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid EndDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if leaveRequest.IsHalfDay {
			endDate = startDate
		}

		days := endDate.Sub(startDate).Hours() / 24

		if leaveRequest.IsHalfDay {
			days = 0.5
		}

		leaveRequest.Days = days
		leaveRequest.StartDate = startDate.Format("2006-01-02")
		leaveRequest.EndDate = endDate.Format("2006-01-02")
		leaveRequest.Status = "Pending"

		if err := db.Create(&leaveRequest).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create leave request"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Send notification email to employee
		err = helper.SendLeaveRequestNotification(employee.Email, leaveRequest.FullNameEmployee, leaveRequest.LeaveType, leaveRequest.StartDate, leaveRequest.EndDate, leaveRequest.Days)
		if err != nil {
			fmt.Println("Failed to send leave request notification:", err)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Leave request created successfully",
			"data":    leaveRequest,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}
*/

func GetAllLeaveRequestsByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
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

		// Searching parameter
		searching := c.QueryParam("searching")

		// Build the query
		query := db.Model(&models.LeaveRequest{}).Where("employee_id = ?", employee.ID)
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where(
				"LOWER(full_name_employee) LIKE ? OR LOWER(leave_type) LIKE ? OR LOWER(start_date) LIKE ? OR LOWER(end_date) LIKE ? OR LOWER(status) LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
			)
		}

		// Retrieve leave requests for the employee with pagination
		var leaveRequests []models.LeaveRequest
		result = query.Order("id DESC").Offset(offset).Limit(perPage).Find(&leaveRequests)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch leave requests"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Batch processing for FullNameEmployee and LeaveType
		var employeeIDs []uint
		employeeMap := make(map[uint]string)
		var leaveTypeIDs []uint
		leaveTypeMap := make(map[uint]string)

		for _, lr := range leaveRequests {
			if _, found := employeeMap[lr.EmployeeID]; !found {
				employeeIDs = append(employeeIDs, lr.EmployeeID)
			}
			if _, found := leaveTypeMap[lr.LeaveTypeID]; !found {
				leaveTypeIDs = append(leaveTypeIDs, lr.LeaveTypeID)
			}
		}

		// Fetch full names for employee IDs
		var employees []models.Employee
		err = db.Model(&models.Employee{}).Where("id IN (?)", employeeIDs).Find(&employees).Error
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching employees"})
		}

		// Create map for fast lookup
		for _, emp := range employees {
			employeeMap[emp.ID] = emp.FullName
		}

		// Fetch leave types for leave type IDs
		var leaveRequestTypes []models.LeaveRequestType
		err = db.Model(&models.LeaveRequestType{}).Where("id IN (?)", leaveTypeIDs).Find(&leaveRequestTypes).Error
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching leave request types"})
		}

		// Create map for fast lookup
		for _, lt := range leaveRequestTypes {
			leaveTypeMap[lt.ID] = lt.LeaveType
		}

		// Update FullNameEmployee and LeaveType fields
		tx := db.Begin()
		for i := range leaveRequests {
			leaveRequests[i].FullNameEmployee = employeeMap[leaveRequests[i].EmployeeID]
			leaveRequests[i].LeaveType = leaveTypeMap[leaveRequests[i].LeaveTypeID]

			if err := tx.Save(&leaveRequests[i]).Error; err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error saving leave requests"})
			}
		}
		tx.Commit()

		// Get total count of leave request records for the employee
		var totalCount int64
		query.Count(&totalCount)

		// Map leave requests to LeaveRequestResponse format
		var leaveRequestsResponse []LeaveRequestResponse
		for _, leaveRequest := range leaveRequests {
			leaveRequestsResponse = append(leaveRequestsResponse, LeaveRequestResponse{
				ID:               leaveRequest.ID,
				EmployeeID:       leaveRequest.EmployeeID,
				Username:         leaveRequest.Username,
				FullNameEmployee: leaveRequest.FullNameEmployee,
				LeaveTypeID:      leaveRequest.LeaveTypeID,
				LeaveType:        leaveRequest.LeaveType,
				StartDate:        leaveRequest.StartDate,
				EndDate:          leaveRequest.EndDate,
				IsHalfDay:        leaveRequest.IsHalfDay,
				Remarks:          leaveRequest.Remarks,
				LeaveReason:      leaveRequest.LeaveReason,
				Days:             leaveRequest.Days,
				Status:           leaveRequest.Status,
				CreatedAt:        leaveRequest.CreatedAt,
				UpdatedAt:        leaveRequest.UpdatedAt,
			})
		}

		// Return the leave request data with pagination info
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave requests fetched successfully",
			"data":    leaveRequestsResponse,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func GetAllLeaveRequestsByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
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

		// Searching parameter
		searching := c.QueryParam("searching")

		// Build the query
		query := db.Model(&models.LeaveRequest{}).Where("employee_id = ?", employee.ID)
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where(
				"LOWER(full_name_employee) LIKE ? OR LOWER(leave_type) LIKE ? OR LOWER(start_date) LIKE ? OR LOWER(end_date) LIKE ? OR LOWER(status) LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
			)
		}

		// Retrieve leave requests for the employee with pagination
		var leaveRequests []models.LeaveRequest
		result = query.Order("id DESC").Offset(offset).Limit(perPage).Find(&leaveRequests)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch leave requests"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Map leave requests to LeaveRequestResponse format
		var leaveRequestsResponse []LeaveRequestResponse
		for _, leaveRequest := range leaveRequests {
			leaveRequestsResponse = append(leaveRequestsResponse, LeaveRequestResponse{
				ID:               leaveRequest.ID,
				EmployeeID:       leaveRequest.EmployeeID,
				Username:         leaveRequest.Username,
				FullNameEmployee: leaveRequest.FullNameEmployee,
				LeaveTypeID:      leaveRequest.LeaveTypeID,
				LeaveType:        leaveRequest.LeaveType,
				StartDate:        leaveRequest.StartDate,
				EndDate:          leaveRequest.EndDate,
				IsHalfDay:        leaveRequest.IsHalfDay,
				Remarks:          leaveRequest.Remarks,
				LeaveReason:      leaveRequest.LeaveReason,
				Days:             leaveRequest.Days,
				Status:           leaveRequest.Status,
				CreatedAt:        leaveRequest.CreatedAt,
				UpdatedAt:        leaveRequest.UpdatedAt,
			})
		}

		// Get total count of leave request records for the employee
		var totalCount int64
		query.Count(&totalCount)

		// Return the leave request data with pagination info
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave requests fetched successfully",
			"data":    leaveRequestsResponse,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

/*
func GetAllLeaveRequestsByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
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

		// Searching parameter
		searching := c.QueryParam("searching")

		// Build the query
		query := db.Model(&models.LeaveRequest{}).Where("employee_id = ?", employee.ID)
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where(
				"LOWER(full_name_employee) LIKE ? OR LOWER(leave_type) LIKE ? OR LOWER(start_date) LIKE ? OR LOWER(end_date) LIKE ? OR LOWER(status) LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
			)
		}

		// Retrieve leave requests for the employee with pagination
		var leaveRequests []models.LeaveRequest
		result = query.Order("id DESC").Offset(offset).Limit(perPage).Find(&leaveRequests)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch leave requests"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Get total count of leave request records for the employee
		var totalCount int64
		query.Count(&totalCount)

		// Return the leave request data with pagination info
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave requests fetched successfully",
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
*/

func GetLeaveRequestByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Retrieve leave request ID from URL parameter
		leaveRequestID := c.Param("id")

		// Retrieve leave request from database
		var leaveRequest models.LeaveRequest
		result = db.First(&leaveRequest, "id = ?", leaveRequestID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Check if the leave request belongs to the logged-in employee
		if leaveRequest.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "You do not have permission to view this leave request"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Map leave request to LeaveRequestResponse format
		response := LeaveRequestResponse{
			ID:               leaveRequest.ID,
			EmployeeID:       leaveRequest.EmployeeID,
			Username:         leaveRequest.Username,
			FullNameEmployee: leaveRequest.FullNameEmployee,
			LeaveTypeID:      leaveRequest.LeaveTypeID,
			LeaveType:        leaveRequest.LeaveType,
			StartDate:        leaveRequest.StartDate,
			EndDate:          leaveRequest.EndDate,
			IsHalfDay:        leaveRequest.IsHalfDay,
			Remarks:          leaveRequest.Remarks,
			LeaveReason:      leaveRequest.LeaveReason,
			Days:             leaveRequest.Days,
			Status:           leaveRequest.Status,
			CreatedAt:        leaveRequest.CreatedAt,
			UpdatedAt:        leaveRequest.UpdatedAt,
		}

		// Return success response with leave request data
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request fetched successfully",
			"data":    response,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func GetLeaveRequestByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		leaveRequestID := c.Param("id")
		var leaveRequest models.LeaveRequest
		result = db.First(&leaveRequest, "id = ?", leaveRequestID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if leaveRequest.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "You do not have permission to view this leave request"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request fetched successfully",
			"data":    leaveRequest,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func UpdateLeaveRequestByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Retrieve leave request ID from URL parameter
		leaveRequestID := c.Param("id")
		if leaveRequestID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Leave request ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve leave request from database
		var leaveRequest models.LeaveRequest
		result = db.First(&leaveRequest, "id = ?", leaveRequestID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Check if the leave request belongs to the logged-in employee
		if leaveRequest.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "You do not have permission to update this leave request"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind updated data from request body
		var updatedLeaveRequest models.LeaveRequest
		if err := c.Bind(&updatedLeaveRequest); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update fields that are allowed to be updated by the employee
		if updatedLeaveRequest.LeaveTypeID != 0 {
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
			startDate, err := time.Parse("2006-01-02", updatedLeaveRequest.StartDate)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid StartDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			leaveRequest.StartDate = startDate.Format("2006-01-02")
		}
		if updatedLeaveRequest.EndDate != "" {
			endDate, err := time.Parse("2006-01-02", updatedLeaveRequest.EndDate)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid EndDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			leaveRequest.EndDate = endDate.Format("2006-01-02")
		}

		if updatedLeaveRequest.IsHalfDay != leaveRequest.IsHalfDay {
			leaveRequest.IsHalfDay = updatedLeaveRequest.IsHalfDay

			if updatedLeaveRequest.IsHalfDay {
				leaveRequest.EndDate = leaveRequest.StartDate
				leaveRequest.Days = 0.5
			} else {
				if leaveRequest.StartDate != "" && leaveRequest.EndDate != "" {
					startDate, _ := time.Parse("2006-01-02", leaveRequest.StartDate)
					endDate, _ := time.Parse("2006-01-02", leaveRequest.EndDate)
					days := endDate.Sub(startDate).Hours() / 24
					leaveRequest.Days = days
				}
			}
		} else {
			if leaveRequest.StartDate != "" && leaveRequest.EndDate != "" {
				startDate, _ := time.Parse("2006-01-02", leaveRequest.StartDate)
				endDate, _ := time.Parse("2006-01-02", leaveRequest.EndDate)
				days := endDate.Sub(startDate).Hours() / 24
				leaveRequest.Days = days
			}
		}

		if updatedLeaveRequest.Remarks != "" {
			leaveRequest.Remarks = updatedLeaveRequest.Remarks
		}
		if updatedLeaveRequest.LeaveReason != "" {
			leaveRequest.LeaveReason = updatedLeaveRequest.LeaveReason
		}

		// Status should not be updated by the employee
		leaveRequest.Status = leaveRequest.Status

		// Save updated leave request to database
		if err := db.Save(&leaveRequest).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update leave request"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Map leave request to LeaveRequestResponse format
		response := LeaveRequestResponse{
			ID:               leaveRequest.ID,
			EmployeeID:       leaveRequest.EmployeeID,
			Username:         leaveRequest.Username,
			FullNameEmployee: leaveRequest.FullNameEmployee,
			LeaveTypeID:      leaveRequest.LeaveTypeID,
			LeaveType:        leaveRequest.LeaveType,
			StartDate:        leaveRequest.StartDate,
			EndDate:          leaveRequest.EndDate,
			IsHalfDay:        leaveRequest.IsHalfDay,
			Remarks:          leaveRequest.Remarks,
			LeaveReason:      leaveRequest.LeaveReason,
			Days:             leaveRequest.Days,
			Status:           leaveRequest.Status,
			CreatedAt:        leaveRequest.CreatedAt,
			UpdatedAt:        leaveRequest.UpdatedAt,
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request updated successfully",
			"data":    response,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func UpdateLeaveRequestByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		leaveRequestID := c.Param("id")
		if leaveRequestID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Leave request ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var leaveRequest models.LeaveRequest
		result = db.First(&leaveRequest, "id = ?", leaveRequestID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if leaveRequest.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "You do not have permission to update this leave request"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		var updatedLeaveRequest models.LeaveRequest
		if err := c.Bind(&updatedLeaveRequest); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update fields that are allowed to be updated by the employee
		if updatedLeaveRequest.LeaveTypeID != 0 {
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
			startDate, err := time.Parse("2006-01-02", updatedLeaveRequest.StartDate)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid StartDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			leaveRequest.StartDate = startDate.Format("2006-01-02")
		}
		if updatedLeaveRequest.EndDate != "" {
			endDate, err := time.Parse("2006-01-02", updatedLeaveRequest.EndDate)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid EndDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			leaveRequest.EndDate = endDate.Format("2006-01-02")
		}

		if updatedLeaveRequest.IsHalfDay != leaveRequest.IsHalfDay {
			leaveRequest.IsHalfDay = updatedLeaveRequest.IsHalfDay

			if updatedLeaveRequest.IsHalfDay {
				leaveRequest.EndDate = leaveRequest.StartDate
				leaveRequest.Days = 0.5
			} else {
				if leaveRequest.StartDate != "" && leaveRequest.EndDate != "" {
					startDate, _ := time.Parse("2006-01-02", leaveRequest.StartDate)
					endDate, _ := time.Parse("2006-01-02", leaveRequest.EndDate)
					days := endDate.Sub(startDate).Hours() / 24
					leaveRequest.Days = days
				}
			}
		} else {
			if leaveRequest.StartDate != "" && leaveRequest.EndDate != "" {
				startDate, _ := time.Parse("2006-01-02", leaveRequest.StartDate)
				endDate, _ := time.Parse("2006-01-02", leaveRequest.EndDate)
				days := endDate.Sub(startDate).Hours() / 24
				leaveRequest.Days = days
			}
		}

		if updatedLeaveRequest.Remarks != "" {
			leaveRequest.Remarks = updatedLeaveRequest.Remarks
		}
		if updatedLeaveRequest.LeaveReason != "" {
			leaveRequest.LeaveReason = updatedLeaveRequest.LeaveReason
		}

		// Status should not be updated by the employee
		leaveRequest.Status = leaveRequest.Status

		if err := db.Save(&leaveRequest).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update leave request"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request updated successfully",
			"data":    leaveRequest,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func DeleteLeaveRequestByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract the token from the request header
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

		// Verify the token and extract the username
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Retrieve the employee details based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Get the leave request ID from the URL parameters
		leaveRequestID := c.Param("id")
		if leaveRequestID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Leave request ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the leave request based on the ID
		var leaveRequest models.LeaveRequest
		result = db.First(&leaveRequest, "id = ?", leaveRequestID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Check if the leave request belongs to the logged-in employee
		if leaveRequest.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "You do not have permission to delete this leave request"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Delete the leave request
		if err := db.Delete(&leaveRequest).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete leave request"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Return a successful response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
