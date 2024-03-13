package controllers

import (
	"fmt"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
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

func GetAllLeaveRequestTypesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Fetch all leave request types from the database
		var leaveRequestTypes []models.LeaveRequestType
		db.Find(&leaveRequestTypes)

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

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":                http.StatusOK,
			"error":               false,
			"message":             "All leave request types retrieved successfully",
			"leave_request_types": leaveRequestTypes,
		})
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
