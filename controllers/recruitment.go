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

		// Validasi data new job
		if newJob.Title == "" || newJob.JobType == "" || newJob.DesignationID == 0 || newJob.NumberOfPosition == 0 ||
			newJob.DateClosing == "" || newJob.Gender == "" || newJob.MinimumExperience == "" ||
			newJob.ShortDescription == "" || newJob.LongDescription == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Incomplete new job data"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Cek apakah designation dengan ID tersebut ada di database
		var designation models.Designation
		result = db.First(&designation, "id = ?", newJob.DesignationID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Designation not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Set designation_name berdasarkan designation_id yang diberikan
		newJob.DesignationName = designation.DesignationName

		// Parse inputan date_closing ke dalam format yang diinginkan (yyyy-mm-dd)
		dateClosing, err := time.Parse("2006-01-02", newJob.DateClosing)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid DateClosing format. Use yyyy-mm-dd format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		newJob.DateClosing = dateClosing.Format("2006-01-02")

		// Set tanggal pembuatan pekerjaan
		newJob.CreatedAt = time.Now()

		// Simpan data new job ke database
		if err := db.Create(&newJob).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create new job"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Berikan respons sukses
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "New job created successfully",
			"data":    newJob,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// GetAllNewJobsByAdmin handles the retrieval of all new jobs by admin with pagination
func GetAllNewJobsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve new jobs from the database with pagination
		var newJobs []models.NewJob
		db.Offset(offset).Limit(perPage).Find(&newJobs)

		// Get total count of new jobs
		var totalCount int64
		db.Model(&models.NewJob{}).Count(&totalCount)

		// Respond with success
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

		// Retrieve new job ID from the request URL parameter
		newJobID := c.Param("id")
		if newJobID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "New job ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch the new job from the database based on the ID
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

		// Retrieve new job ID from the request URL parameter
		newJobID := c.Param("id")
		if newJobID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "New job ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch the new job from the database based on the ID
		var newJob models.NewJob
		result = db.First(&newJob, "id = ?", newJobID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "New job not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated new job data from the request body
		var updatedJob models.NewJob
		if err := c.Bind(&updatedJob); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update the new job data
		if updatedJob.Title != "" {
			newJob.Title = updatedJob.Title
		}
		if updatedJob.JobType != "" {
			newJob.JobType = updatedJob.JobType
		}
		if updatedJob.DesignationID != 0 {
			// Fetch the designation data based on the new designation ID
			var designation models.Designation
			result = db.First(&designation, "id = ?", updatedJob.DesignationID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Designation not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			newJob.DesignationID = updatedJob.DesignationID
			newJob.DesignationName = designation.DesignationName // Update designation name accordingly
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
		if updatedJob.Gender != "" {
			newJob.Gender = updatedJob.Gender
		}
		if updatedJob.MinimumExperience != "" {
			newJob.MinimumExperience = updatedJob.MinimumExperience
		}
		if updatedJob.ShortDescription != "" {
			newJob.ShortDescription = updatedJob.ShortDescription
		}
		if updatedJob.LongDescription != "" {
			newJob.LongDescription = updatedJob.LongDescription
		}

		// Update the new job in the database
		db.Save(&newJob)

		// Provide success response
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

		// Retrieve new job ID from the request URL parameter
		newJobID := c.Param("id")
		if newJobID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "New job ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch the new job from the database based on the ID
		var newJob models.NewJob
		result = db.First(&newJob, "id = ?", newJobID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "New job not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the new job from the database
		db.Delete(&newJob)

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "New job deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
