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

func CreateExitStatusByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var exitStatus models.Exit
		if err := c.Bind(&exitStatus); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if exitStatus.ExitName == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Exit status name is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate ExitName
		if len(exitStatus.ExitName) < 5 || len(exitStatus.ExitName) > 30 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Exit name must be between 5 and 30 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingExitStatus models.Exit
		result = db.Where("exit_name = ?", exitStatus.ExitName).First(&existingExitStatus)
		if result.Error == nil {
			errorResponse := helper.Response{Code: http.StatusConflict, Error: true, Message: "Exit status with this name already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		}

		currentTime := time.Now()
		exitStatus.CreatedAt = &currentTime

		db.Create(&exitStatus)

		successResponse := helper.Response{
			Code:    http.StatusCreated,
			Error:   false,
			Message: "Exit status created successfully",
			Exit:    &exitStatus,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllExitStatusByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var exitStatuses []models.Exit
		var totalCount int64
		db.Model(&models.Exit{}).Count(&totalCount)
		db.Order("id DESC").Offset(offset).Limit(perPage).Find(&exitStatuses)

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

func GetExitStatusByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		exitID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid exit status ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var exitStatus models.Exit
		result = db.First(&exitStatus, exitID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Exit status not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

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

		exitID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid exit ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var exitStatus models.Exit
		result = db.First(&exitStatus, exitID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Exit status not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedExitStatus models.Exit
		if err := c.Bind(&updatedExitStatus); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedExitStatus.ExitName == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Exit status name is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedExitStatus.ExitName != "" {
			if len(updatedExitStatus.ExitName) < 5 || len(updatedExitStatus.ExitName) > 30 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Exit name must be between 5 and 30 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			exitStatus.ExitName = updatedExitStatus.ExitName
		}

		if updatedExitStatus.ExitName != exitStatus.ExitName {
			var existingExitStatus models.Exit
			result = db.Where("exit_name = ?", updatedExitStatus.ExitName).First(&existingExitStatus)
			if result.Error == nil {
				errorResponse := helper.Response{Code: http.StatusConflict, Error: true, Message: "Exit status with this name already exists"}
				return c.JSON(http.StatusConflict, errorResponse)
			}
		}

		db.Model(&exitStatus).Updates(&updatedExitStatus)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Exit status updated successfully",
			Exit:    &exitStatus,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteExitStatusByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		exitID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid exit status ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var exitStatus models.Exit
		result = db.First(&exitStatus, exitID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Exit status not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&exitStatus)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Exit status deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
