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

// CreateHelpdeskByEmployee memungkinkan karyawan untuk menambahkan data helpdesk
func CreateHelpdeskByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the helpdesk data from the request body
		var helpdesk models.Helpdesk
		if err := c.Bind(&helpdesk); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate helpdesk data
		if helpdesk.Subject == "" || helpdesk.Priority == "" || helpdesk.Description == "" || helpdesk.DepartmentID == 0 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "All fields are required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Set the employee ID and username
		helpdesk.EmployeeID = employee.ID
		helpdesk.EmployeeUsername = employee.Username
		helpdesk.EmployeeFullName = employee.FirstName + " " + employee.LastName

		// Check if the department with the given ID exists
		var existingDepartment models.Department
		result = db.First(&existingDepartment, helpdesk.DepartmentID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Set the DepartmentName
		helpdesk.DepartmentName = existingDepartment.DepartmentName

		// Set the created timestamp
		currentTime := time.Now()
		helpdesk.CreatedAt = &currentTime

		// Create the helpdesk in the database
		db.Create(&helpdesk)

		// Respond with success
		successResponse := helper.Response{
			Code:     http.StatusCreated,
			Error:    false,
			Message:  "Helpdesk created successfully",
			Helpdesk: &helpdesk,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// GetHelpdeskByEmployee memungkinkan karyawan untuk melihat semua data helpdesk miliknya
func GetAllHelpdeskByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Retrieve all helpdesk data for the employee
		var helpdeskList []models.Helpdesk
		db.Where("employee_id = ?", employee.ID).Find(&helpdeskList)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"message":  "Helpdesk data retrieved successfully",
			"helpdesk": helpdeskList,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetHelpdeskByIDForEmployee memungkinkan karyawan untuk melihat data helpdesk miliknya berdasarkan ID helpdesk
func GetHelpdeskByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Extract ID parameter from the request
		id := c.Param("id")

		// Parse ID to uint
		helpdeskID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the helpdesk data from the database by ID
		var helpdesk models.Helpdesk
		result = db.First(&helpdesk, helpdeskID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Helpdesk not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Check if the helpdesk belongs to the employee
		if helpdesk.EmployeeID != employee.ID {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Helpdesk does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"message":  "Helpdesk data retrieved successfully",
			"helpdesk": helpdesk,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// UpdateHelpdeskByIDForEmployee memungkinkan karyawan untuk mengedit data helpdesk miliknya berdasarkan ID helpdesk
func UpdateHelpdeskByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Extract helpdesk ID from the request params
		helpdeskIDStr := c.Param("id")
		helpdeskID, err := strconv.ParseUint(helpdeskIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid helpdesk ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve helpdesk data from the database based on ID
		var helpdesk models.Helpdesk
		result = db.First(&helpdesk, uint(helpdeskID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Helpdesk data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Check if the helpdesk belongs to the employee
		if helpdesk.EmployeeID != employee.ID {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Helpdesk does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind updated helpdesk data from the request body
		var updatedHelpdesk models.Helpdesk
		if err := c.Bind(&updatedHelpdesk); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update helpdesk data selectively
		if updatedHelpdesk.Subject != "" {
			helpdesk.Subject = updatedHelpdesk.Subject
		}

		if updatedHelpdesk.DepartmentID != 0 {
			// Check if the department with the given ID exists
			var existingDepartment models.Department
			result = db.First(&existingDepartment, updatedHelpdesk.DepartmentID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			// Set the DepartmentName
			helpdesk.DepartmentID = updatedHelpdesk.DepartmentID
			helpdesk.DepartmentName = existingDepartment.DepartmentName
		}

		if updatedHelpdesk.Priority != "" {
			helpdesk.Priority = updatedHelpdesk.Priority
		}

		if updatedHelpdesk.Description != "" {
			helpdesk.Description = updatedHelpdesk.Description
		}

		// Set the updated timestamp
		currentTime := time.Now()
		helpdesk.UpdatedAt = currentTime

		db.Save(&helpdesk)

		// Respond with success
		successResponse := helper.Response{
			Code:     http.StatusOK,
			Error:    false,
			Message:  "Helpdesk data updated successfully",
			Helpdesk: &helpdesk,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// DeleteHelpdeskByIDForEmployee memungkinkan karyawan untuk menghapus data helpdesk miliknya berdasarkan ID helpdesk
func DeleteHelpdeskByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Extract helpdesk ID from the request params
		helpdeskIDStr := c.Param("id")
		helpdeskID, err := strconv.ParseUint(helpdeskIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid helpdesk ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve helpdesk data from the database based on ID
		var helpdesk models.Helpdesk
		result = db.First(&helpdesk, uint(helpdeskID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Helpdesk data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Check if the helpdesk belongs to the employee
		if helpdesk.EmployeeID != employee.ID {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Helpdesk does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Delete the helpdesk from the database
		db.Delete(&helpdesk)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Helpdesk deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
