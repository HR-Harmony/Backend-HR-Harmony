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

func CreateAnnouncementByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var announcement models.Announcement
		if err := c.Bind(&announcement); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check each required field individually and return specific error messages
		if announcement.Title == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Title is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		if announcement.Summary == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Summary is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		if announcement.Description == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Description is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		if announcement.StartDate == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "StartDate is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		if announcement.EndDate == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "EndDate is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		startDate, err := time.Parse("2006-01-02", announcement.StartDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid StartDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		endDate, err := time.Parse("2006-01-02", announcement.EndDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid EndDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		announcement.StartDate = startDate.Format("2006-01-02")
		announcement.EndDate = endDate.Format("2006-01-02")

		var existingDepartment models.Department
		result = db.First(&existingDepartment, announcement.DepartmentID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		announcement.DepartmentName = existingDepartment.DepartmentName

		currentTime := time.Now()
		announcement.CreatedAt = &currentTime

		db.Create(&announcement)

		successResponse := helper.Response{
			Code:         http.StatusCreated,
			Error:        false,
			Message:      "Announcement created successfully",
			Announcement: &announcement,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

/*
func CreateAnnouncementByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var announcement models.Announcement
		if err := c.Bind(&announcement); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if announcement.Title == "" || announcement.Summary == "" || announcement.Description == "" || announcement.StartDate == "" || announcement.EndDate == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "All fields are required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		startDate, err := time.Parse("2006-01-02", announcement.StartDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid StartDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		endDate, err := time.Parse("2006-01-02", announcement.EndDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid EndDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		announcement.StartDate = startDate.Format("2006-01-02")
		announcement.EndDate = endDate.Format("2006-01-02")

		var existingDepartment models.Department
		result = db.First(&existingDepartment, announcement.DepartmentID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		announcement.DepartmentName = existingDepartment.DepartmentName

		currentTime := time.Now()
		announcement.CreatedAt = &currentTime

		db.Create(&announcement)

		successResponse := helper.Response{
			Code:         http.StatusCreated,
			Error:        false,
			Message:      "Announcement created successfully",
			Announcement: &announcement,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}
*/

func GetAnnouncementsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		searching := c.QueryParam("searching")

		var announcements []models.Announcement
		query := db.Order("id DESC").Offset(offset).Limit(perPage)

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where(
				db.Where("title ILIKE ?", searchPattern).
					Or("department_name ILIKE ?", searchPattern).
					Or("description ILIKE ?", searchPattern).
					Or("start_date::text ILIKE ?", searchPattern).
					Or("end_date::text ILIKE ?", searchPattern))
		}

		if err := query.Find(&announcements).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Error fetching announcements"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var totalCount int64
		db.Model(&models.Announcement{}).Count(&totalCount)

		successResponse := map[string]interface{}{
			"code":          http.StatusOK,
			"error":         false,
			"message":       "Announcements retrieved successfully",
			"announcements": announcements,
			"pagination":    map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}

		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAnnouncementByIDForAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		announcementID := c.Param("id")

		var announcement models.Announcement
		result = db.First(&announcement, announcementID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Announcement not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := helper.Response{
			Code:         http.StatusOK,
			Error:        false,
			Message:      "Announcement retrieved successfully",
			Announcement: &announcement,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateAnnouncementForAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		announcementID := c.Param("id")

		var announcement models.Announcement
		result = db.First(&announcement, announcementID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Announcement not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedAnnouncement models.Announcement
		if err := c.Bind(&updatedAnnouncement); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedAnnouncement.Title != "" {
			announcement.Title = updatedAnnouncement.Title
		}

		if updatedAnnouncement.DepartmentID != 0 {
			var newDepartment models.Department
			result = db.First(&newDepartment, updatedAnnouncement.DepartmentID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "New department not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			announcement.DepartmentID = updatedAnnouncement.DepartmentID
			announcement.DepartmentName = newDepartment.DepartmentName
		}

		if updatedAnnouncement.Summary != "" {
			announcement.Summary = updatedAnnouncement.Summary
		}

		if updatedAnnouncement.Description != "" {
			announcement.Description = updatedAnnouncement.Description
		}

		if updatedAnnouncement.StartDate != "" {
			startDate, err := time.Parse("2006-01-02", updatedAnnouncement.StartDate)
			if err != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid StartDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			announcement.StartDate = startDate.Format("2006-01-02")
		}

		if updatedAnnouncement.EndDate != "" {
			endDate, err := time.Parse("2006-01-02", updatedAnnouncement.EndDate)
			if err != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid EndDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			announcement.EndDate = endDate.Format("2006-01-02")
		}

		db.Save(&announcement)

		successResponse := helper.Response{
			Code:         http.StatusOK,
			Error:        false,
			Message:      "Announcement updated successfully",
			Announcement: &announcement,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteAnnouncementForAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		announcementID := c.Param("id")

		var announcement models.Announcement
		result = db.First(&announcement, announcementID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Announcement not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&announcement)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Announcement deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
