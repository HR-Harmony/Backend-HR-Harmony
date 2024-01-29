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

func CreateHelpdeskByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Bind the helpdesk data from the request body
		var helpdesk models.Helpdesk
		if err := c.Bind(&helpdesk); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate helpdesk data
		if helpdesk.Subject == "" || helpdesk.Priority == "" || helpdesk.Description == "" || helpdesk.DepartmentID == 0 || helpdesk.EmployeeID == 0 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "All fields are required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the department with the given ID exists
		var existingDepartment models.Department
		result = db.First(&existingDepartment, helpdesk.DepartmentID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Check if the employee with the given ID exists
		var existingEmployee models.Employee
		result = db.First(&existingEmployee, helpdesk.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Set the DepartmentName and EmployeeUsername
		helpdesk.DepartmentName = existingDepartment.DepartmentName
		helpdesk.EmployeeUsername = existingEmployee.Username

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

func GetAllHelpdeskByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve all helpdesk data from the database
		var helpdesks []models.Helpdesk
		db.Find(&helpdesks)

		// Respond with success
		successResponse := helper.Response{
			Code:      http.StatusOK,
			Error:     false,
			Message:   "All helpdesk data retrieved successfully",
			Helpdesks: helpdesks,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetHelpdeskByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Respond with success
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

		if updatedHelpdesk.Priority != "" {
			helpdesk.Priority = updatedHelpdesk.Priority
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

		if updatedHelpdesk.EmployeeID != 0 {
			// Check if the employee with the given ID exists
			var existingEmployee models.Employee
			result = db.First(&existingEmployee, updatedHelpdesk.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			// Set the EmployeeUsername
			helpdesk.EmployeeID = updatedHelpdesk.EmployeeID
			helpdesk.EmployeeUsername = existingEmployee.Username
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

func DeleteHelpdeskByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Delete helpdesk data from the database
		db.Delete(&helpdesk)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Helpdesk data deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
