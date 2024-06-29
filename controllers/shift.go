package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func CreateShiftByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var shift models.Shift
		if err := c.Bind(&shift); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if shift.ShiftName == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Shift name is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate RoleName using regexp
		if len(shift.ShiftName) < 5 || len(shift.ShiftName) > 30 || !regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(shift.ShiftName) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Shift name must be between 5 and 30 characters and contain only letters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingShift models.Shift
		result = db.Where("shift_name = ?", shift.ShiftName).First(&existingShift)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Shift with this name already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		}

		currentTime := time.Now()
		shift.CreatedAt = &currentTime

		db.Create(&shift)

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

		// Handle search parameter
		searching := c.QueryParam("searching")

		var shifts []models.Shift
		query := db.Order("id DESC").Offset(offset).Limit(perPage)

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("shift_name ILIKE ?", searchPattern)
		}

		if err := query.Find(&shifts).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Error fetching shifts"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var totalCount int64
		db.Model(&models.Shift{}).Count(&totalCount)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Shifts retrieved successfully",
			"shifts":  shifts,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetShiftByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		shiftIDStr := c.Param("id")
		shiftID, err := strconv.ParseUint(shiftIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid shift ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var shift models.Shift
		result = db.First(&shift, uint(shiftID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Shift not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Shift retrieved successfully",
			Shift:   &shift,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func EditShiftByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		shiftID := c.Param("id")
		if shiftID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Shift ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var shift models.Shift
		result = db.First(&shift, "id = ?", shiftID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Shift not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedShift models.Shift
		if err := c.Bind(&updatedShift); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		/*
			if updatedShift.ShiftName != "" {
				shift.ShiftName = updatedShift.ShiftName
			}
		*/

		if updatedShift.ShiftName != "" {
			if len(updatedShift.ShiftName) < 5 || len(updatedShift.ShiftName) > 30 || !regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(updatedShift.ShiftName) {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Shift name must be between 5 and 30 characters and can only be letters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
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

		db.Save(&shift)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Shift updated successfully",
			"data":    shift,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteShiftByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		shiftIDStr := c.Param("id")
		shiftID, err := strconv.ParseUint(shiftIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid shift ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var shift models.Shift
		result = db.First(&shift, uint(shiftID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Shift not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&shift)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Shift deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
