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

		var leaveRequestType models.LeaveRequestType
		if err := c.Bind(&leaveRequestType); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if leaveRequestType.LeaveType == "" || leaveRequestType.DaysPerYears <= 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Incomplete leave request type data"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if err := db.Create(&leaveRequestType).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create leave request type"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Leave request type created successfully",
			"data":    leaveRequestType,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllLeaveRequestTypesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var leaveRequestTypes []models.LeaveRequestType
		db.Model(&models.LeaveRequestType{}).Order("id DESC").Offset(offset).Limit(perPage).Find(&leaveRequestTypes)

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

		leaveRequestTypeID := c.Param("id")
		if leaveRequestTypeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Leave request type ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

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

		leaveRequestTypeID := c.Param("id")
		if leaveRequestTypeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Leave request type ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var leaveRequestType models.LeaveRequestType
		result = db.First(&leaveRequestType, "id = ?", leaveRequestTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave request type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedLeaveRequestType models.LeaveRequestType
		if err := c.Bind(&updatedLeaveRequestType); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedLeaveRequestType.LeaveType != "" {
			leaveRequestType.LeaveType = updatedLeaveRequestType.LeaveType
		}
		if updatedLeaveRequestType.DaysPerYears != 0 {
			leaveRequestType.DaysPerYears = updatedLeaveRequestType.DaysPerYears
		}
		if updatedLeaveRequestType.IsRequiresApproval != leaveRequestType.IsRequiresApproval {
			leaveRequestType.IsRequiresApproval = updatedLeaveRequestType.IsRequiresApproval
		}

		db.Save(&leaveRequestType)

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

		leaveRequestTypeID := c.Param("id")
		if leaveRequestTypeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Leave request type ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var leaveRequestType models.LeaveRequestType
		result = db.First(&leaveRequestType, "id = ?", leaveRequestTypeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Leave request type not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Check if the leave request type is associated with any leave requests
		var leaveRequestCount int64
		db.Model(&models.LeaveRequest{}).Where("leave_type_id = ?", leaveRequestTypeID).Count(&leaveRequestCount)
		if leaveRequestCount > 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Cannot delete leave request type because it is associated with one or more leave requests"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		db.Delete(&leaveRequestType)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request type deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

type LeaveRequestResponse struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	EmployeeID       uint       `json:"employee_id"`
	Username         string     `json:"username"`
	FullNameEmployee string     `json:"full_name_employee"`
	LeaveTypeID      uint       `json:"leave_type_id"`
	LeaveType        string     `json:"leave_type"`
	StartDate        string     `json:"start_date"`
	EndDate          string     `json:"end_date"`
	IsHalfDay        bool       `json:"is_half_day"`
	Remarks          string     `json:"remarks"`
	LeaveReason      string     `json:"leave_reason"`
	Days             float64    `json:"days"`
	Status           string     `json:"status"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func CreateLeaveRequestByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var leaveRequest models.LeaveRequest
		if err := c.Bind(&leaveRequest); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var employee models.Employee
		result = db.First(&employee, "id = ?", leaveRequest.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
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
func CreateLeaveRequestByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var leaveRequest models.LeaveRequest
		if err := c.Bind(&leaveRequest); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var employee models.Employee
		result = db.First(&employee, "id = ?", leaveRequest.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
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

func GetAllLeaveRequestsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		searching := c.QueryParam("searching")

		var totalCount int64
		query := db.Model(&models.LeaveRequest{})
		if searching != "" {
			searching = strings.ToLower(searching)
			query = query.Where("LOWER(full_name_employee) LIKE ? OR LOWER(username) LIKE ? OR LOWER(leave_type) LIKE ? OR LOWER(status) LIKE ?",
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
			)
		}
		query.Count(&totalCount)

		/*
			// Fetch leave requests with preloaded employee and leave request type data
			var leaveRequests []models.LeaveRequest
			err = query.Preload("Employee").Preload("LeaveRequestType").Order("id DESC").Offset(offset).Limit(perPage).Find(&leaveRequests).Error
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching leave requests"})
			}
		*/

		// Fetch leave requests with preloaded employee and leave request type data
		var leaveRequests []models.LeaveRequest
		err = query.Order("id DESC").Offset(offset).Limit(perPage).Find(&leaveRequests).Error
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching leave requests"})
		}

		// Batch processing for FullNameEmployee
		// Collect unique employee IDs
		var employeeIDs []uint
		employeeMap := make(map[uint]string)
		for _, lr := range leaveRequests {
			if _, found := employeeMap[lr.EmployeeID]; !found {
				employeeIDs = append(employeeIDs, lr.EmployeeID)
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

		// Update FullNameEmployee field and save to database
		tx := db.Begin()
		for i := range leaveRequests {
			leaveRequests[i].FullNameEmployee = employeeMap[leaveRequests[i].EmployeeID]
			if err := tx.Save(&leaveRequests[i]).Error; err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error saving leave requests"})
			}
		}

		// Batch processing for LeaveType
		// Collect unique leave type IDs
		var leaveTypeIDs []uint
		leaveTypeMap := make(map[uint]string)
		for _, lr := range leaveRequests {
			if _, found := leaveTypeMap[lr.LeaveTypeID]; !found {
				leaveTypeIDs = append(leaveTypeIDs, lr.LeaveTypeID)
			}
		}

		// Fetch leave types for leave type IDs
		var leaveRequestTypes []models.LeaveRequestType
		err = db.Model(&models.LeaveRequestType{}).Where("id IN (?)", leaveTypeIDs).Find(&leaveRequestTypes).Error
		if err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching leave request types"})
		}

		// Create map for fast lookup
		for _, lt := range leaveRequestTypes {
			leaveTypeMap[lt.ID] = lt.LeaveType
		}

		// Update LeaveType field and save to database
		for i := range leaveRequests {
			leaveRequests[i].LeaveType = leaveTypeMap[leaveRequests[i].LeaveTypeID]
			if err := tx.Save(&leaveRequests[i]).Error; err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error saving leave requests"})
			}
		}

		tx.Commit()

		// Prepare response with LeaveRequestResponse structure
		var leaveRequestResponses []LeaveRequestResponse
		for _, lr := range leaveRequests {
			leaveRequestResponse := LeaveRequestResponse{
				ID:               lr.ID,
				EmployeeID:       lr.EmployeeID,
				Username:         lr.Username,
				FullNameEmployee: lr.FullNameEmployee,
				LeaveTypeID:      lr.LeaveTypeID,
				LeaveType:        lr.LeaveType,
				StartDate:        lr.StartDate,
				EndDate:          lr.EndDate,
				IsHalfDay:        lr.IsHalfDay,
				Remarks:          lr.Remarks,
				LeaveReason:      lr.LeaveReason,
				Days:             lr.Days,
				Status:           lr.Status,
				CreatedAt:        lr.CreatedAt,
				UpdatedAt:        lr.UpdatedAt,
			}
			leaveRequestResponses = append(leaveRequestResponses, leaveRequestResponse)
		}

		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Leave request data retrieved successfully",
			"data":       leaveRequestResponses,
			"pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func GetAllLeaveRequestsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		searching := c.QueryParam("searching")

		var totalCount int64
		query := db.Model(&models.LeaveRequest{})
		if searching != "" {
			searching = strings.ToLower(searching)
			query = query.Where("LOWER(full_name_employee) LIKE ? OR LOWER(username) LIKE ? OR LOWER(leave_type) LIKE ? OR LOWER(status) LIKE ?",
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
			)
		}
		query.Count(&totalCount)

		// Fetch leave requests with preloaded employee and leave request type data
		var leaveRequests []models.LeaveRequest
		err = query.Preload("Employee").Preload("LeaveRequestType").Order("id DESC").Offset(offset).Limit(perPage).Find(&leaveRequests).Error
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching leave requests"})
		}

		// Batch processing for FullNameEmployee
		// Collect unique employee IDs
		var employeeIDs []uint
		employeeMap := make(map[uint]string)
		for _, lr := range leaveRequests {
			if _, found := employeeMap[lr.EmployeeID]; !found {
				employeeIDs = append(employeeIDs, lr.EmployeeID)
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

		// Update FullNameEmployee field
		for i := range leaveRequests {
			leaveRequests[i].FullNameEmployee = employeeMap[leaveRequests[i].EmployeeID]
		}

		// Batch processing for LeaveType
		// Collect unique leave type IDs
		var leaveTypeIDs []uint
		leaveTypeMap := make(map[uint]string)
		for _, lr := range leaveRequests {
			if _, found := leaveTypeMap[lr.LeaveTypeID]; !found {
				leaveTypeIDs = append(leaveTypeIDs, lr.LeaveTypeID)
			}
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

		// Update LeaveType field
		for i := range leaveRequests {
			leaveRequests[i].LeaveType = leaveTypeMap[leaveRequests[i].LeaveTypeID]
		}

		// Prepare response with LeaveRequestResponse structure
		var leaveRequestResponses []LeaveRequestResponse
		for _, lr := range leaveRequests {
			leaveRequestResponse := LeaveRequestResponse{
				ID:               lr.ID,
				EmployeeID:       lr.EmployeeID,
				Username:         lr.Username,
				FullNameEmployee: lr.FullNameEmployee,
				LeaveTypeID:      lr.LeaveTypeID,
				LeaveType:        lr.LeaveType,
				StartDate:        lr.StartDate,
				EndDate:          lr.EndDate,
				IsHalfDay:        lr.IsHalfDay,
				Remarks:          lr.Remarks,
				LeaveReason:      lr.LeaveReason,
				Days:             lr.Days,
				Status:           lr.Status,
				CreatedAt:        lr.CreatedAt,
				UpdatedAt:        lr.UpdatedAt,
			}
			leaveRequestResponses = append(leaveRequestResponses, leaveRequestResponse)
		}

		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Leave request data retrieved successfully",
			"data":       leaveRequestResponses,
			"pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

/*
func GetAllLeaveRequestsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		searching := c.QueryParam("searching")

		var leaveRequests []models.LeaveRequest
		query := db.Model(&models.LeaveRequest{})
		if searching != "" {
			searching = strings.ToLower(searching)
			query = query.Where("LOWER(full_name_employee) LIKE ? OR LOWER(username) LIKE ? OR LOWER(leave_type) LIKE ? OR LOWER(status) LIKE ?",
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
			)
		}
		query.Order("id DESC").Offset(offset).Limit(perPage).Find(&leaveRequests)

		var totalCount int64
		db.Model(&models.LeaveRequest{}).Count(&totalCount)

		// Prepare response with LeaveRequestResponse structure
		var leaveRequestResponses []LeaveRequestResponse
		for _, lr := range leaveRequests {
			leaveRequestResponse := LeaveRequestResponse{
				ID:               lr.ID,
				EmployeeID:       lr.EmployeeID,
				Username:         lr.Username,
				FullNameEmployee: lr.FullNameEmployee,
				LeaveTypeID:      lr.LeaveTypeID,
				LeaveType:        lr.LeaveType,
				StartDate:        lr.StartDate,
				EndDate:          lr.EndDate,
				IsHalfDay:        lr.IsHalfDay,
				Remarks:          lr.Remarks,
				LeaveReason:      lr.LeaveReason,
				Days:             lr.Days,
				Status:           lr.Status,
				CreatedAt:        lr.CreatedAt,
				UpdatedAt:        lr.UpdatedAt,
			}
			leaveRequestResponses = append(leaveRequestResponses, leaveRequestResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request data retrieved successfully",
			"data":    leaveRequestResponses,
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
func GetAllLeaveRequestsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		searching := c.QueryParam("searching")

		var leaveRequests []models.LeaveRequest
		query := db.Model(&models.LeaveRequest{})
		if searching != "" {
			searching = strings.ToLower(searching)
			query = query.Where("LOWER(full_name_employee) LIKE ? OR LOWER(username) LIKE ? OR LOWER(leave_type) LIKE ? OR LOWER(status) LIKE ?",
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
			)
		}
		query.Order("id DESC").Offset(offset).Limit(perPage).Find(&leaveRequests)

		var totalCount int64
		db.Model(&models.LeaveRequest{}).Count(&totalCount)

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
*/

/*
func GetAllLeaveRequestsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		searching := c.QueryParam("searching")

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
		db.Order("id DESC").Offset(offset).Limit(perPage).Find(&leaveRequests)

		var totalCount int64
		db.Model(&models.LeaveRequest{}).Count(&totalCount)

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
*/

func GetLeaveRequestByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			"message": "Leave request retrieved successfully",
			"data":    response,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func GetLeaveRequestByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request retrieved successfully",
			"data":    leaveRequest,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func UpdateLeaveRequestByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var updatedLeaveRequest models.LeaveRequest
		if err := c.Bind(&updatedLeaveRequest); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedLeaveRequest.EmployeeID != 0 {
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

		oldStatus := leaveRequest.Status
		if updatedLeaveRequest.Status != "" {
			leaveRequest.Status = updatedLeaveRequest.Status
		}

		if err := db.Save(&leaveRequest).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update leave request"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Retrieve related employee data from the database using the EmployeeID in leaveRequest
		var employee models.Employee
		result = db.First(&employee, "id = ?", leaveRequest.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Call email notification function if status has changed and employee email is available
		if updatedLeaveRequest.Status != "" && employee.Email != "" {
			err = helper.SendLeaveRequestStatusNotification(employee.Email, leaveRequest.FullNameEmployee, oldStatus, updatedLeaveRequest.Status)
			if err != nil {
				fmt.Println("Failed to send leave request status notification:", err)
			}
		}

		// Prepare response with LeaveRequestResponse structure
		leaveRequestResponse := LeaveRequestResponse{
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
			"data":    leaveRequestResponse,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func UpdateLeaveRequestByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var updatedLeaveRequest models.LeaveRequest
		if err := c.Bind(&updatedLeaveRequest); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedLeaveRequest.EmployeeID != 0 {
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

		oldStatus := leaveRequest.Status

		if updatedLeaveRequest.Status != "" {
			leaveRequest.Status = updatedLeaveRequest.Status
		}

		if err := db.Save(&leaveRequest).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update leave request"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Dapatkan data karyawan terkait dari basis data menggunakan EmployeeID yang ada di leaveRequest
		var employee models.Employee
		result = db.First(&employee, "id = ?", leaveRequest.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Panggil fungsi notifikasi email jika status berubah dan email employee tersedia
		if updatedLeaveRequest.Status != "" && employee.Email != "" {
			err = helper.SendLeaveRequestStatusNotification(employee.Email, leaveRequest.FullNameEmployee, oldStatus, updatedLeaveRequest.Status)
			if err != nil {
				fmt.Println("Failed to send leave request status notification:", err)
			}
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

func DeleteLeaveRequestByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		if err := db.Delete(&leaveRequest).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete leave request"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Leave request deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
