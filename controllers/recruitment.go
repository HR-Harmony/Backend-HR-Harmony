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

func CreateNewJobByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var newJob models.NewJob
		if err := c.Bind(&newJob); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if newJob.Title == "" || newJob.JobType == "" || newJob.DesignationID == 0 || newJob.NumberOfPosition == 0 ||
			newJob.DateClosing == "" || newJob.MinimumExperience == "" ||
			newJob.ShortDescription == "" || newJob.LongDescription == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Incomplete new job data"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(newJob.Title) < 5 || len(newJob.Title) > 100 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Title must be between 5 and 100 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(newJob.ShortDescription) < 5 || len(newJob.ShortDescription) > 300 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Short description must be between 5 and 300 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(newJob.LongDescription) < 5 || len(newJob.LongDescription) > 3000 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "long description must be between 5 and 3000 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var designation models.Designation
		result = db.First(&designation, "id = ?", newJob.DesignationID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Designation not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		newJob.DesignationName = designation.DesignationName

		dateClosing, err := time.Parse("2006-01-02", newJob.DateClosing)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid DateClosing format. Use yyyy-mm-dd format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		newJob.DateClosing = dateClosing.Format("2006-01-02")

		newJob.CreatedAt = time.Now()

		if err := db.Create(&newJob).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create new job"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "New job created successfully",
			"data":    newJob,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllNewJobsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		var newJobs []models.NewJob
		db.Order("id DESC").Offset(offset).Limit(perPage).Find(&newJobs)

		var totalCount int64
		db.Model(&models.NewJob{}).Count(&totalCount)

		successResponse := map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"message":  "All new jobs retrieved successfully",
			"new_jobs": newJobs,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetNewJobByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		newJobID := c.Param("id")
		if newJobID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "New job ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var newJob models.NewJob
		result = db.First(&newJob, "id = ?", newJobID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "New job not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "New job retrieved successfully",
			"new_job": newJob,
		})
	}
}

func UpdateNewJobByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		newJobID := c.Param("id")
		if newJobID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "New job ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var newJob models.NewJob
		result = db.First(&newJob, "id = ?", newJobID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "New job not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedJob models.NewJob
		if err := c.Bind(&updatedJob); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		/*
			if updatedJob.Title != "" {
				newJob.Title = updatedJob.Title
			}
		*/

		if updatedJob.Title != "" {
			if len(updatedJob.Title) < 5 || len(updatedJob.Title) > 100 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Title must be between 5 and 100"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			newJob.Title = updatedJob.Title
		}

		if updatedJob.JobType != "" {
			newJob.JobType = updatedJob.JobType
		}
		if updatedJob.DesignationID != 0 {
			var designation models.Designation
			result = db.First(&designation, "id = ?", updatedJob.DesignationID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Designation not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			newJob.DesignationID = updatedJob.DesignationID
			newJob.DesignationName = designation.DesignationName
		}
		if updatedJob.NumberOfPosition != 0 {
			newJob.NumberOfPosition = updatedJob.NumberOfPosition
		}
		if updatedJob.IsPublish != newJob.IsPublish {
			newJob.IsPublish = updatedJob.IsPublish
		}
		if updatedJob.DateClosing != "" {
			newJob.DateClosing = updatedJob.DateClosing
		}
		if updatedJob.MinimumExperience != "" {
			newJob.MinimumExperience = updatedJob.MinimumExperience
		}

		/*
			if updatedJob.ShortDescription != "" {
				newJob.ShortDescription = updatedJob.ShortDescription
			}
		*/
		if updatedJob.ShortDescription != "" {
			if len(updatedJob.ShortDescription) < 5 || len(updatedJob.ShortDescription) > 300 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Short description must be between 5 and 300"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			newJob.ShortDescription = updatedJob.ShortDescription
		}

		/*
			if updatedJob.LongDescription != "" {
				newJob.LongDescription = updatedJob.LongDescription
			}
		*/

		if updatedJob.LongDescription != "" {
			if len(updatedJob.LongDescription) < 5 || len(updatedJob.LongDescription) > 3000 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Long description must be between 5 and 3000"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			newJob.LongDescription = updatedJob.LongDescription
		}

		db.Save(&newJob)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "New job updated successfully",
			"data":    newJob,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteNewJobByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		newJobID := c.Param("id")
		if newJobID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "New job ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var newJob models.NewJob
		result = db.First(&newJob, "id = ?", newJobID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "New job not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&newJob)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "New job deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
