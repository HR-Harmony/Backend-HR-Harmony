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

// CreateDesignationByAdmin handles the creation of a new designation by admin
func CreateDesignationByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Bind the designation data from the request body
		var designation models.Designation
		if err := c.Bind(&designation); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate designation data
		if designation.DesignationName == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Designation name is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if designation.DepartmentID == 0 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Department ID is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the department with the given ID exists
		var existingDepartment models.Department
		result = db.Where("id = ?", designation.DepartmentID).First(&existingDepartment)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Set the created timestamp
		currentTime := time.Now()
		designation.CreatedAt = currentTime

		// Create the designation in the database
		db.Create(&designation)

		// Respond with success
		successResponse := helper.Response{
			Code:        http.StatusCreated,
			Error:       false,
			Message:     "Designation created successfully",
			Designation: &designation,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// GetDesignationsByAdmin handles the retrieval of all designations by admin
func GetAllDesignationsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve all designations from the database
		var designations []models.Designation
		db.Find(&designations)

		// Respond with success
		successResponse := helper.Response{
			Code:         http.StatusOK,
			Error:        false,
			Message:      "Designations retrieved successfully",
			Designations: designations,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetDesignationByID handles the retrieval of a designation by admin based on its ID
func GetDesignationByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get the designation ID from the request parameters
		designationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid designation ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the designation from the database based on its ID
		var designation models.Designation
		result = db.Where("id = ?", designationID).First(&designation)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Designation not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with success
		successResponse := helper.Response{
			Code:        http.StatusOK,
			Error:       false,
			Message:     "Designation retrieved successfully",
			Designation: &designation,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// UpdateDesignationByID handles the update of a designation by ID
func UpdateDesignationByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get the designation ID from the request parameters
		designationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid designation ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the existing designation from the database based on its ID
		var existingDesignation models.Designation
		result = db.Where("id = ?", designationID).First(&existingDesignation)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Designation not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated designation data from the request body
		var updatedDesignation models.Designation
		if err := c.Bind(&updatedDesignation); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate the updated designation data
		if updatedDesignation.DesignationName == "" && updatedDesignation.Description == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "At least one field (designation_name or description) is required for update"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update only the fields that are provided in the request
		if updatedDesignation.DesignationName != "" {
			existingDesignation.DesignationName = updatedDesignation.DesignationName
		}

		if updatedDesignation.Description != "" {
			existingDesignation.Description = updatedDesignation.Description
		}

		// Save the changes to the database
		db.Save(&existingDesignation)

		// Respond with success
		successResponse := helper.Response{
			Code:        http.StatusOK,
			Error:       false,
			Message:     "Designation updated successfully",
			Designation: &existingDesignation,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// DeleteDesignationByID handles the deletion of a designation by admin based on its ID
func DeleteDesignationByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get the designation ID from the request parameters
		designationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid designation ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the existing designation from the database based on its ID
		var existingDesignation models.Designation
		result = db.Where("id = ?", designationID).First(&existingDesignation)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Designation not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the designation from the database
		db.Delete(&existingDesignation)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Designation deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
