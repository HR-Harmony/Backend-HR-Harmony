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

func CreateHelpdeskByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var helpdesk models.Helpdesk
		if err := c.Bind(&helpdesk); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if helpdesk.Subject == "" || helpdesk.Priority == "" || helpdesk.Description == "" || helpdesk.DepartmentID == 0 || helpdesk.EmployeeID == 0 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "All fields are required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingDepartment models.Department
		result = db.First(&existingDepartment, helpdesk.DepartmentID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var existingEmployee models.Employee
		result = db.First(&existingEmployee, helpdesk.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		helpdesk.DepartmentName = existingDepartment.DepartmentName
		helpdesk.EmployeeUsername = existingEmployee.Username
		helpdesk.EmployeeFullName = existingEmployee.FirstName + " " + existingEmployee.LastName

		helpdesk.Status = "Open"

		currentTime := time.Now()
		helpdesk.CreatedAt = &currentTime

		if err := db.Create(&helpdesk).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to create helpdesk"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Send notification email to employee
		err = helper.SendHelpdeskNotification(existingEmployee.Email, helpdesk.EmployeeFullName, helpdesk.Subject, helpdesk.Description)
		if err != nil {
			fmt.Println("Failed to send helpdesk notification:", err)
		}

		successResponse := helper.Response{
			Code:     http.StatusCreated,
			Error:    false,
			Message:  "Helpdesk created successfully",
			Helpdesk: &helpdesk,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// GetAllHelpdeskByAdmin handles the retrieval of all helpdesk data by admin with pagination
func GetAllHelpdeskByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var helpdesks []models.Helpdesk
		var totalCount int64
		db.Model(&models.Helpdesk{}).Count(&totalCount)
		db.Order("id DESC").Offset(offset).Limit(perPage).Find(&helpdesks)

		successResponse := map[string]interface{}{
			"code":      http.StatusOK,
			"error":     false,
			"message":   "Helpdesk data retrieved successfully",
			"helpdesks": helpdesks,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetHelpdeskByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		helpdeskIDStr := c.Param("id")
		helpdeskID, err := strconv.ParseUint(helpdeskIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid helpdesk ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var helpdesk models.Helpdesk
		result = db.First(&helpdesk, uint(helpdeskID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Helpdesk data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := helper.Response{
			Code:     http.StatusOK,
			Error:    false,
			Message:  "Helpdesk data retrieved successfully",
			Helpdesk: &helpdesk,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateHelpdeskByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		helpdeskIDStr := c.Param("id")
		helpdeskID, err := strconv.ParseUint(helpdeskIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid helpdesk ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var helpdesk models.Helpdesk
		result = db.First(&helpdesk, uint(helpdeskID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Helpdesk data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedHelpdesk models.Helpdesk
		if err := c.Bind(&updatedHelpdesk); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedHelpdesk.Subject != "" {
			helpdesk.Subject = updatedHelpdesk.Subject
		}

		if updatedHelpdesk.Priority != "" {
			helpdesk.Priority = updatedHelpdesk.Priority
		}

		if updatedHelpdesk.DepartmentID != 0 {
			var existingDepartment models.Department
			result = db.First(&existingDepartment, updatedHelpdesk.DepartmentID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			helpdesk.DepartmentID = updatedHelpdesk.DepartmentID
			helpdesk.DepartmentName = existingDepartment.DepartmentName
		}

		if updatedHelpdesk.EmployeeID != 0 {
			var existingEmployee models.Employee
			result = db.First(&existingEmployee, updatedHelpdesk.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			helpdesk.EmployeeID = updatedHelpdesk.EmployeeID
			helpdesk.EmployeeUsername = existingEmployee.Username
			helpdesk.EmployeeFullName = existingEmployee.FirstName + " " + existingEmployee.LastName
		}

		if updatedHelpdesk.Description != "" {
			helpdesk.Description = updatedHelpdesk.Description
		}

		if updatedHelpdesk.Status != "" {
			helpdesk.Status = updatedHelpdesk.Status
		}

		currentTime := time.Now()
		helpdesk.UpdatedAt = currentTime

		if err := db.Save(&helpdesk).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to update helpdesk"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Dapatkan data karyawan terkait dari basis data menggunakan EmployeeID yang ada di helpdesk
		var employee models.Employee
		result = db.First(&employee, "id = ?", helpdesk.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Panggil fungsi notifikasi email jika status berubah dan email employee tersedia
		if updatedHelpdesk.Status != "" && employee.Email != "" {
			err = helper.SendHelpdeskNotificationStatus(employee.Email, helpdesk.EmployeeFullName, helpdesk.Subject, helpdesk.Description, helpdesk.Status)
			if err != nil {
				fmt.Println("Failed to send helpdesk status notification:", err)
			}
		}

		successResponse := helper.Response{
			Code:     http.StatusOK,
			Error:    false,
			Message:  "Helpdesk data updated successfully",
			Helpdesk: &helpdesk,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteHelpdeskByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		helpdeskIDStr := c.Param("id")
		helpdeskID, err := strconv.ParseUint(helpdeskIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid helpdesk ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var helpdesk models.Helpdesk
		result = db.First(&helpdesk, uint(helpdeskID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Helpdesk data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&helpdesk)
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Helpdesk data deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetTicketStatsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Initialize the ticket status and priority counts
		ticketStatus := map[string]int{
			"Open":   0,
			"Closed": 0,
		}

		ticketPriority := map[string]int{
			"Low":      0,
			"Medium":   0,
			"High":     0,
			"Critical": 0,
		}

		// Query the ticket counts by status
		var ticketStatusCounts []struct {
			Status string
			Count  int
		}
		if err := db.Model(&models.Helpdesk{}).Select("status, count(*) as count").Group("status").Scan(&ticketStatusCounts).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve ticket counts by status"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Update the ticket status counts based on the query results
		for _, count := range ticketStatusCounts {
			statusKey := strings.ReplaceAll(count.Status, " ", "_")
			ticketStatus[statusKey] = count.Count
		}

		// Query the ticket counts by priority
		var ticketPriorityCounts []struct {
			Priority string
			Count    int
		}
		if err := db.Model(&models.Helpdesk{}).Select("priority, count(*) as count").Group("priority").Scan(&ticketPriorityCounts).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve ticket counts by priority"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Update the ticket priority counts based on the query results
		for _, count := range ticketPriorityCounts {
			priorityKey := strings.ReplaceAll(count.Priority, " ", "_")
			ticketPriority[priorityKey] = count.Count
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":            http.StatusOK,
			"error":           false,
			"message":         "Ticket counts by status and priority retrieved successfully",
			"ticket_status":   ticketStatus,
			"ticket_priority": ticketPriority,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
