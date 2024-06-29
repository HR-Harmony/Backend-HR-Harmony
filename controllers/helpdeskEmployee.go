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

func CreateHelpdeskByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var helpdesk models.Helpdesk
		if err := c.Bind(&helpdesk); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if helpdesk.Subject == "" || helpdesk.Priority == "" || helpdesk.Description == "" || helpdesk.DepartmentID == 0 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "All fields are required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(helpdesk.Subject) < 5 || len(helpdesk.Subject) > 30 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Helpdesk subject must be between 5 and 30 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(helpdesk.Description) < 5 || len(helpdesk.Description) > 3000 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Helpdesk description must be between 5 and 3000 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		helpdesk.EmployeeID = employee.ID
		helpdesk.EmployeeUsername = employee.Username
		helpdesk.EmployeeFullName = employee.FirstName + " " + employee.LastName

		var existingDepartment models.Department
		result = db.First(&existingDepartment, helpdesk.DepartmentID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		helpdesk.DepartmentName = existingDepartment.DepartmentName
		helpdesk.Status = "Open"

		currentTime := time.Now()
		helpdesk.CreatedAt = &currentTime

		if err := db.Create(&helpdesk).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to create helpdesk"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Send notification email to employee
		err = helper.SendHelpdeskNotification(employee.Email, helpdesk.EmployeeFullName, helpdesk.Subject, helpdesk.Description)
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

func GetAllHelpdeskByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var helpdeskList []models.Helpdesk
		db.Where("employee_id = ?", employee.ID).Find(&helpdeskList).Order("id DESC")

		successResponse := map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"message":  "Helpdesk data retrieved successfully",
			"helpdesk": helpdeskList,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetHelpdeskByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		id := c.Param("id")

		helpdeskID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var helpdesk models.Helpdesk
		result = db.First(&helpdesk, helpdeskID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Helpdesk not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if helpdesk.EmployeeID != employee.ID {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Helpdesk does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"message":  "Helpdesk data retrieved successfully",
			"helpdesk": helpdesk,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateHelpdeskByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
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

		if helpdesk.EmployeeID != employee.ID {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Helpdesk does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		var updatedHelpdesk models.Helpdesk
		if err := c.Bind(&updatedHelpdesk); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		/*
			if updatedHelpdesk.Subject != "" {
				helpdesk.Subject = updatedHelpdesk.Subject
			}
		*/

		if updatedHelpdesk.Subject != "" {
			if len(updatedHelpdesk.Subject) < 5 || len(updatedHelpdesk.Subject) > 30 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Helpdesk subject must be between 5 and 30 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			helpdesk.Subject = updatedHelpdesk.Subject
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

		if updatedHelpdesk.Priority != "" {
			helpdesk.Priority = updatedHelpdesk.Priority
		}

		/*
			if updatedHelpdesk.Description != "" {
				helpdesk.Description = updatedHelpdesk.Description
			}
		*/

		if updatedHelpdesk.Description != "" {
			if len(updatedHelpdesk.Description) < 5 || len(updatedHelpdesk.Description) > 3000 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Helpdesk description must be between 5 and 3000 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			helpdesk.Description = updatedHelpdesk.Description
		}

		updatedHelpdesk.Status = helpdesk.Status

		currentTime := time.Now()
		helpdesk.UpdatedAt = currentTime

		db.Save(&helpdesk)

		successResponse := helper.Response{
			Code:     http.StatusOK,
			Error:    false,
			Message:  "Helpdesk data updated successfully",
			Helpdesk: &helpdesk,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteHelpdeskByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
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

		if helpdesk.EmployeeID != employee.ID {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Helpdesk does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		db.Delete(&helpdesk)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Helpdesk deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetHelpdeskStatsByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
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

		// Query the ticket counts by status for the specific employee
		var ticketStatusCounts []struct {
			Status string
			Count  int
		}
		if err := db.Model(&models.Helpdesk{}).
			Select("status, count(*) as count").
			Where("employee_id = ?", employee.ID).
			Group("status").
			Scan(&ticketStatusCounts).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve ticket counts by status"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Update the ticket status counts based on the query results
		for _, count := range ticketStatusCounts {
			statusKey := strings.ReplaceAll(count.Status, " ", "_")
			ticketStatus[statusKey] = count.Count
		}

		// Query the ticket counts by priority for the specific employee
		var ticketPriorityCounts []struct {
			Priority string
			Count    int
		}
		if err := db.Model(&models.Helpdesk{}).
			Select("priority, count(*) as count").
			Where("employee_id = ?", employee.ID).
			Group("priority").
			Scan(&ticketPriorityCounts).Error; err != nil {
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
