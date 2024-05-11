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

// CreateExitStatusByAdmin handles the creation of a new exit status by admin
func CreateExitStatusByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Bind the exit status data from the request body
		var exitStatus models.Exit
		if err := c.Bind(&exitStatus); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate exit status data
		if exitStatus.ExitName == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Exit status name is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the exit status name already exists
		var existingExitStatus models.Exit
		result = db.Where("exit_name = ?", exitStatus.ExitName).First(&existingExitStatus)
		if result.Error == nil {
			errorResponse := helper.Response{Code: http.StatusConflict, Error: true, Message: "Exit status with this name already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		}

		// Set the created timestamp
		currentTime := time.Now()
		exitStatus.CreatedAt = &currentTime

		// Create the exit status in the database
		db.Create(&exitStatus)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusCreated,
			Error:   false,
			Message: "Exit status created successfully",
			Exit:    &exitStatus,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// GetAllExitStatusByAdmin handles the retrieval of all exit statuses by admin with pagination
func GetAllExitStatusByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve all exit statuses from the database with pagination
		var exitStatuses []models.Exit
		var totalCount int64
		db.Model(&models.Exit{}).Count(&totalCount)
		db.Offset(offset).Limit(perPage).Find(&exitStatuses)

		// Respond with the list of exit statuses
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Exit statuses retrieved successfully",
			"exits":   exitStatuses,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetExitStatusByIDByAdmin handles the retrieval of an exit status by its ID by admin
func GetExitStatusByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get exit status ID from the request parameter
		exitID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid exit status ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the exit status from the database by ID
		var exitStatus models.Exit
		result = db.First(&exitStatus, exitID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Exit status not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with the exit status
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Exit status retrieved successfully",
			Exit:    &exitStatus,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateExitStatusByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Extract exit ID from the request parameters
		exitID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid exit ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the exit status with the given ID exists
		var exitStatus models.Exit
		result = db.First(&exitStatus, exitID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Exit status not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated exit status data from the request body
		var updatedExitStatus models.Exit
		if err := c.Bind(&updatedExitStatus); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate exit status data (you can customize the validation rules)
		if updatedExitStatus.ExitName == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Exit status name is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the updated exit status name already exists
		if updatedExitStatus.ExitName != exitStatus.ExitName {
			var existingExitStatus models.Exit
			result = db.Where("exit_name = ?", updatedExitStatus.ExitName).First(&existingExitStatus)
			if result.Error == nil {
				errorResponse := helper.Response{Code: http.StatusConflict, Error: true, Message: "Exit status with this name already exists"}
				return c.JSON(http.StatusConflict, errorResponse)
			}
		}

		// Update the exit status in the database
		db.Model(&exitStatus).Updates(&updatedExitStatus)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Exit status updated successfully",
			Exit:    &exitStatus,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// DeleteExitStatusByIDByAdmin handles the deletion of an exit status by its ID by admin
func DeleteExitStatusByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get exit status ID from the request parameter
		exitID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid exit status ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the exit status exists
		var exitStatus models.Exit
		result = db.First(&exitStatus, exitID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Exit status not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the exit status from the database
		db.Delete(&exitStatus)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Exit status deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
