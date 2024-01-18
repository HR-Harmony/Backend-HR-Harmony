package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"strings"
	"time"
)

// CreateAnnouncementByAdmin handles the creation of a new announcement by admin
func CreateAnnouncementByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Bind the announcement data from the request body
		var announcement models.Announcement
		if err := c.Bind(&announcement); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate announcement data
		if announcement.Title == "" || announcement.Summary == "" || announcement.Description == "" || announcement.StartDate == "" || announcement.EndDate == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "All fields are required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Parse start date from string to time.Time
		startDate, err := time.Parse("2006-01-02", announcement.StartDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid StartDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Parse end date from string to time.Time
		endDate, err := time.Parse("2006-01-02", announcement.EndDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid EndDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Format start date and end date in "yyyy-mm-dd" format
		announcement.StartDate = startDate.Format("2006-01-02")
		announcement.EndDate = endDate.Format("2006-01-02")

		// Check if the department with the given ID exists
		var existingDepartment models.Department
		result = db.First(&existingDepartment, announcement.DepartmentID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		announcement.DepartmentName = existingDepartment.DepartmentName

		// Set the created timestamp
		currentTime := time.Now()
		announcement.CreatedAt = &currentTime

		// Create the announcement in the database
		db.Create(&announcement)

		// Respond with success
		successResponse := helper.Response{
			Code:         http.StatusCreated,
			Error:        false,
			Message:      "Announcement created successfully",
			Announcement: &announcement,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// GetAnnouncementsByAdmin retrieves all announcements for admin
func GetAnnouncementsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve all announcements
		var announcements []models.Announcement
		result = db.Find(&announcements)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Error fetching announcements"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Respond with success
		successResponse := helper.Response{
			Code:          http.StatusOK,
			Error:         false,
			Message:       "Announcements retrieved successfully",
			Announcements: announcements,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetAnnouncementByIDForAdmin retrieves announcement data by ID for admin
func GetAnnouncementByIDForAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Extract announcement ID from the request
		announcementID := c.Param("id")

		// Retrieve announcement by ID
		var announcement models.Announcement
		result = db.First(&announcement, announcementID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Announcement not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with success
		successResponse := helper.Response{
			Code:         http.StatusOK,
			Error:        false,
			Message:      "Announcement retrieved successfully",
			Announcement: &announcement,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// UpdateAnnouncementForAdmin handles the update of an announcement by admin
func UpdateAnnouncementForAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Extract announcement ID from the request
		announcementID := c.Param("id")

		// Retrieve announcement by ID
		var announcement models.Announcement
		result = db.First(&announcement, announcementID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Announcement not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated announcement data from the request body
		var updatedAnnouncement models.Announcement
		if err := c.Bind(&updatedAnnouncement); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate and update the announcement fields
		if updatedAnnouncement.Title != "" {
			announcement.Title = updatedAnnouncement.Title
		}

		if updatedAnnouncement.DepartmentID != 0 {
			// Check if the new department ID exists
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
			// Parse and update start date
			startDate, err := time.Parse("2006-01-02", updatedAnnouncement.StartDate)
			if err != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid StartDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			announcement.StartDate = startDate.Format("2006-01-02")
		}

		if updatedAnnouncement.EndDate != "" {
			// Parse and update end date
			endDate, err := time.Parse("2006-01-02", updatedAnnouncement.EndDate)
			if err != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid EndDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			announcement.EndDate = endDate.Format("2006-01-02")
		}

		// Update the announcement in the database
		db.Save(&announcement)

		// Respond with success
		successResponse := helper.Response{
			Code:         http.StatusOK,
			Error:        false,
			Message:      "Announcement updated successfully",
			Announcement: &announcement,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// DeleteAnnouncementForAdmin handles the deletion of an announcement by admin
func DeleteAnnouncementForAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Extract announcement ID from the request
		announcementID := c.Param("id")

		// Retrieve announcement by ID
		var announcement models.Announcement
		result = db.First(&announcement, announcementID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Announcement not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the announcement from the database
		db.Delete(&announcement)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Announcement deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
