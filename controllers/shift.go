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

func CreateShiftByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Bind the shift data from the request body
		var shift models.Shift
		if err := c.Bind(&shift); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate shift data
		if shift.ShiftName == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Shift name is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the shift name already exists
		var existingShift models.Shift
		result = db.Where("shift_name = ?", shift.ShiftName).First(&existingShift)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Shift with this name already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		}

		// Set the created timestamp
		currentTime := time.Now()
		shift.CreatedAt = &currentTime

		// Create the shift in the database
		db.Create(&shift)

		// Respond with success
		successResponse := helper.ResponseShift{
			Code:    http.StatusCreated,
			Error:   false,
			Message: "Shift created successfully",
			Shift:   &shift,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllShiftsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve all shifts from the database
		var shifts []models.Shift
		db.Find(&shifts)

		// Respond with the list of shifts
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Shifts retrieved successfully",
			Shifts:  shifts,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetShiftByIDByAdmin handles the retrieval of a shift by its ID for admin
func GetShiftByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve shift ID from the URL parameter
		shiftIDStr := c.Param("id")
		shiftID, err := strconv.ParseUint(shiftIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid shift ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the shift from the database
		var shift models.Shift
		result = db.First(&shift, uint(shiftID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Shift not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with the shift details
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Shift retrieved successfully",
			Shift:   &shift,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// EditShiftNameByIDByAdmin handles the editing of a shift's shift_name by its ID for admin
func EditShiftByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Extract shift ID from the URL parameter
		shiftID := c.Param("id")
		if shiftID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Shift ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Find the shift by ID
		var shift models.Shift
		result = db.First(&shift, "id = ?", shiftID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Shift not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated shift data from the request body
		var updatedShift models.Shift
		if err := c.Bind(&updatedShift); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update fields that are allowed to be changed
		if updatedShift.ShiftName != "" {
			shift.ShiftName = updatedShift.ShiftName
		}
		if updatedShift.MondayInTime != "" {
			shift.MondayInTime = updatedShift.MondayInTime
		}
		if updatedShift.MondayOutTime != "" {
			shift.MondayOutTime = updatedShift.MondayOutTime
		}
		if updatedShift.TuesdayInTime != "" {
			shift.TuesdayInTime = updatedShift.TuesdayInTime
		}
		if updatedShift.TuesdayOutTime != "" {
			shift.TuesdayOutTime = updatedShift.TuesdayOutTime
		}
		if updatedShift.WednesdayInTime != "" {
			shift.WednesdayInTime = updatedShift.WednesdayInTime
		}
		if updatedShift.WednesdayOutTime != "" {
			shift.WednesdayOutTime = updatedShift.WednesdayOutTime
		}
		if updatedShift.ThursdayInTime != "" {
			shift.ThursdayInTime = updatedShift.ThursdayInTime
		}
		if updatedShift.ThursdayOutTime != "" {
			shift.ThursdayOutTime = updatedShift.ThursdayOutTime
		}
		if updatedShift.FridayInTime != "" {
			shift.FridayInTime = updatedShift.FridayInTime
		}
		if updatedShift.FridayOutTime != "" {
			shift.FridayOutTime = updatedShift.FridayOutTime
		}
		if updatedShift.SaturdayInTime != "" {
			shift.SaturdayInTime = updatedShift.SaturdayInTime
		}
		if updatedShift.SaturdayOutTime != "" {
			shift.SaturdayOutTime = updatedShift.SaturdayOutTime
		}
		if updatedShift.SundayInTime != "" {
			shift.SundayInTime = updatedShift.SundayInTime
		}
		if updatedShift.SundayOutTime != "" {
			shift.SundayOutTime = updatedShift.SundayOutTime
		}

		// Save the changes to the database
		db.Save(&shift)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Shift updated successfully",
			"data":    shift,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// DeleteShiftByIDByAdmin handles the deletion of a shift by its ID for admin
func DeleteShiftByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve shift ID from the URL parameter
		shiftIDStr := c.Param("id")
		shiftID, err := strconv.ParseUint(shiftIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid shift ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the shift from the database
		var shift models.Shift
		result = db.First(&shift, uint(shiftID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Shift not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the shift from the database
		db.Delete(&shift)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Shift deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
