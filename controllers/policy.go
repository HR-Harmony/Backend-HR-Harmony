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

func CreatePolicyByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var policy models.Policy
		if err := c.Bind(&policy); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if policy.Title == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Policy title is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		currentTime := time.Now()
		policy.CreatedAt = &currentTime

		policy.CreatedByAdminID = adminUser.ID
		policy.CreatedByAdminUsername = adminUser.Username

		db.Create(&policy)

		successResponse := helper.Response{
			Code:    http.StatusCreated,
			Error:   false,
			Message: "Policy created successfully",
			Policy:  &policy,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllPoliciesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Handle search parameters
		searching := c.QueryParam("searching")

		var policies []models.Policy
		query := db.Order("id DESC").Offset(offset).Limit(perPage)

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("title ILIKE ?", searchPattern)
		}

		if err := query.Find(&policies).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch Policy records"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var totalCount int64
		countQuery := db.Model(&models.Policy{})
		if searching != "" {
			searchPattern := "%" + searching + "%"
			countQuery = countQuery.Where("title ILIKE ?", searchPattern)
		}
		countQuery.Count(&totalCount)

		successResponse := map[string]interface{}{
			"Code":       http.StatusOK,
			"Error":      false,
			"Message":    "Policies retrieved successfully",
			"Policies":   policies,
			"Pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetPolicyByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		policyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid policy ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var policy models.Policy
		result = db.First(&policy, policyID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Policy not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Policy retrieved successfully",
			Policy:  &policy,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdatePolicyByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		policyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid policy ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingPolicy models.Policy
		result = db.First(&existingPolicy, policyID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Policy not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedPolicy models.Policy
		if err := c.Bind(&updatedPolicy); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedPolicy.Title != "" {
			existingPolicy.Title = updatedPolicy.Title
		}

		if updatedPolicy.Description != "" {
			existingPolicy.Description = updatedPolicy.Description
		}

		db.Save(&existingPolicy)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Policy updated successfully",
			Policy:  &existingPolicy,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeletePolicyByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		policyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid policy ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingPolicy models.Policy
		result = db.First(&existingPolicy, policyID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Policy not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&existingPolicy)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Policy deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
