// controllers/createRole.go

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

func CreateRoleByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var role models.Role
		if err := c.Bind(&role); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if role.RoleName == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Role name is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate RoleName using regexp
		if len(role.RoleName) < 5 || len(role.RoleName) > 30 || !regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(role.RoleName) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Role name must be between 5 and 30 characters and contain only letters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingRole models.Role
		result = db.Where("role_name = ?", role.RoleName).First(&existingRole)
		if result.Error == nil {
			errorResponse := helper.Response{Code: http.StatusConflict, Error: true, Message: "Role with this name already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		}

		currentTime := time.Now()
		role.CreatedAt = &currentTime

		db.Create(&role)

		successResponse := helper.Response{
			Code:    http.StatusCreated,
			Error:   false,
			Message: "Role created successfully",
			Role:    &role,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllRolesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var roles []models.Role
		query := db.Order("id DESC").Offset(offset).Limit(perPage)

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("role_name ILIKE ?", searchPattern)
		}

		if err := query.Find(&roles).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Error fetching roles"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var totalCount int64
		db.Model(&models.Role{}).Count(&totalCount)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Roles retrieved successfully",
			"roles":   roles,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetRoleByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		roleIDStr := c.Param("id")
		roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid role ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var role models.Role
		result = db.First(&role, uint(roleID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Role not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Message: "Role retrieved successfully",
			Error:   false,
			Role:    &role,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func EditRoleByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		roleIDStr := c.Param("id")
		roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid role ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var role models.Role
		result = db.First(&role, uint(roleID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Role not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedRole models.Role
		if err := c.Bind(&updatedRole); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedRole.RoleName == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Role name is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate RoleName using regexp
		if len(updatedRole.RoleName) < 5 || len(updatedRole.RoleName) > 30 || !regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(updatedRole.RoleName) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Role name must be between 5 and 30 characters and contain only letters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		role.RoleName = updatedRole.RoleName
		role.UpdatedAt = time.Now()

		db.Save(&role)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Role updated successfully",
			Role:    &role,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteRoleByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		roleIDStr := c.Param("id")
		roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid role ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var role models.Role
		result = db.First(&role, uint(roleID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Role not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&role)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Role deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
