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

func CreateTrainerByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind the trainer data from the request body
		var trainer models.Trainer
		if err := c.Bind(&trainer); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Combine first name and last name to create full name
		trainer.FullName = trainer.FirstName + " " + trainer.LastName

		// Save the trainer to the database
		if err := db.Create(&trainer).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create trainer"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Trainer created successfully",
			"data":    trainer,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllTrainersByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Fetch searching query parameter
		searching := c.QueryParam("searching")

		// Fetch trainers data from database with optional search filters
		var trainers []models.Trainer
		query := db.Model(&trainers)
		if searching != "" {
			query = query.Where("LOWER(full_name) LIKE ? OR contact_number LIKE ? OR LOWER(email) LIKE ? OR LOWER(expertise) LIKE ?",
				"%"+strings.ToLower(searching)+"%",
				"%"+searching+"%",
				"%"+strings.ToLower(searching)+"%",
				"%"+strings.ToLower(searching)+"%",
			)
		}
		query.Find(&trainers)

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Trainers fetched successfully",
			"data":    trainers,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetTrainerByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Retrieve trainer ID from the request URL parameter
		trainerID := c.Param("id")
		if trainerID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Trainer ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch the trainer from the database based on the ID
		var trainer models.Trainer
		result = db.First(&trainer, "id = ?", trainerID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Trainer not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Trainer fetched successfully",
			"data":    trainer,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateTrainerByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Retrieve trainer ID from the request URL parameter
		trainerID := c.Param("id")
		if trainerID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Trainer ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch the trainer from the database based on the ID
		var trainer models.Trainer
		result = db.First(&trainer, "id = ?", trainerID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Trainer not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated trainer data from the request body
		var updatedTrainer models.Trainer
		if err := c.Bind(&updatedTrainer); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update the trainer data
		if updatedTrainer.FirstName != "" {
			trainer.FirstName = updatedTrainer.FirstName
			trainer.FullName = trainer.FirstName + " " + trainer.LastName // Update full name
		}
		if updatedTrainer.LastName != "" {
			trainer.LastName = updatedTrainer.LastName
			trainer.FullName = trainer.FirstName + " " + trainer.LastName // Update full name
		}
		if updatedTrainer.ContactNumber != "" {
			trainer.ContactNumber = updatedTrainer.ContactNumber
		}
		if updatedTrainer.Email != "" {
			trainer.Email = updatedTrainer.Email
		}
		if updatedTrainer.Expertise != "" {
			trainer.Expertise = updatedTrainer.Expertise
		}
		if updatedTrainer.Address != "" {
			trainer.Address = updatedTrainer.Address
		}

		// Update the trainer in the database
		db.Save(&trainer)

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Trainer updated successfully",
			"data":    trainer,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteTrainerByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Retrieve trainer ID from the request URL parameter
		trainerID := c.Param("id")
		if trainerID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Trainer ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch the trainer from the database based on the ID
		var trainer models.Trainer
		result = db.First(&trainer, "id = ?", trainerID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Trainer not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the trainer from the database
		db.Delete(&trainer)

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Trainer deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func CreateTrainingSkillByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind the TrainingSkill data from the request body
		var trainingSkill models.TrainingSkill
		if err := c.Bind(&trainingSkill); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Save the TrainingSkill to the database
		if err := db.Create(&trainingSkill).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create TrainingSkill"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "TrainingSkill created successfully",
			"data":    trainingSkill,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllTrainingSkillsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Fetch searching query parameter
		searching := strings.ToLower(c.QueryParam("searching"))

		// Fetch TrainingSkills data from database with optional search filter
		var trainingSkills []models.TrainingSkill
		query := db.Model(&trainingSkills)
		if searching != "" {
			query = query.Where("LOWER(training_skill) LIKE ?", "%"+searching+"%")
		}
		query.Find(&trainingSkills)

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "TrainingSkills fetched successfully",
			"data":    trainingSkills,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetTrainingSkillByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Get ID from path parameter
		id := c.Param("id")

		// Fetch TrainingSkill data by ID
		var trainingSkill models.TrainingSkill
		if err := db.First(&trainingSkill, "id = ?", id).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "TrainingSkill not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "TrainingSkill fetched successfully",
			"data":    trainingSkill,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateTrainingSkillByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Get ID from path parameter
		id := c.Param("id")

		// Bind the updated TrainingSkill data from the request body
		var updatedTrainingSkill models.TrainingSkill
		if err := c.Bind(&updatedTrainingSkill); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch TrainingSkill data by ID
		var trainingSkill models.TrainingSkill
		if err := db.First(&trainingSkill, "id = ?", id).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "TrainingSkill not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Update the TrainingSkill data
		if err := db.Model(&trainingSkill).Updates(updatedTrainingSkill).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update TrainingSkill"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "TrainingSkill updated successfully",
			"data":    trainingSkill,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteTrainingSkillByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Get ID from path parameter
		id := c.Param("id")

		// Fetch TrainingSkill data by ID
		var trainingSkill models.TrainingSkill
		if err := db.First(&trainingSkill, "id = ?", id).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "TrainingSkill not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the TrainingSkill data
		if err := db.Delete(&trainingSkill).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete TrainingSkill"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "TrainingSkill deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func CreateTrainingByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind the training data from the request body
		var training models.Training
		if err := c.Bind(&training); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate date format for start_date
		_, err = time.Parse("2006-01-02", training.StartDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid start_date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate date format for end_date
		_, err = time.Parse("2006-01-02", training.EndDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid end_date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch trainer data based on TrainerID
		var trainer models.Trainer
		result = db.First(&trainer, "id = ?", training.TrainerID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Trainer not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		training.FullNameTrainer = trainer.FullName

		// Fetch training skill data based on TrainingSkillID
		var trainingSkill models.TrainingSkill
		result = db.First(&trainingSkill, "id = ?", training.TrainingSkillID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Training skill not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		training.TrainingSkill = trainingSkill.TrainingSkill

		// Fetch employee data based on EmployeeID
		var employee models.Employee
		result = db.First(&employee, "id = ?", training.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		training.FullNameEmployee = employee.FullName

		// Save the training data to the database
		if err := db.Create(&training).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create training"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Training created successfully",
			"data":    training,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllTrainingsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Fetch searching query parameter
		searching := c.QueryParam("searching")

		// Fetch trainings data from database with optional search filters
		var trainings []models.Training
		query := db.Model(&trainings)
		if searching != "" {
			// Convert searching string to lowercase
			searching = strings.ToLower(searching)
			query = query.Where("LOWER(full_name_trainer) LIKE ? OR LOWER(training_skill) LIKE ? OR LOWER(full_name_employee) LIKE ? OR CAST(training_cost AS VARCHAR) LIKE ?",
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
				searching,
			)
		}
		query.Find(&trainings)
		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Trainings fetched successfully",
			"data":    trainings,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetTrainingByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Extract training ID from request parameters
		trainingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid training ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch training data by ID
		var training models.Training
		result = db.First(&training, trainingID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Training not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Training fetched successfully",
			"data":    training,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// UpdateTrainingByID mengizinkan hanya admin untuk mengedit data pelatihan berdasarkan ID.
func UpdateTrainingByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Extract training ID from request parameters
		trainingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid training ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch training data by ID
		var training models.Training
		result = db.First(&training, trainingID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Training not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated data from the request body
		var updatedTraining models.Training
		if err := c.Bind(&updatedTraining); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update the training fields
		if updatedTraining.TrainerID != 0 {
			// Fetch trainer data based on TrainerID
			var trainer models.Trainer
			result = db.First(&trainer, "id = ?", updatedTraining.TrainerID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Trainer not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			training.TrainerID = updatedTraining.TrainerID
			training.FullNameTrainer = trainer.FullName
		}
		if updatedTraining.TrainingSkillID != 0 {
			// Fetch training skill data based on TrainingSkillID
			var trainingSkill models.TrainingSkill
			result = db.First(&trainingSkill, "id = ?", updatedTraining.TrainingSkillID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Training skill not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			training.TrainingSkillID = updatedTraining.TrainingSkillID
			training.TrainingSkill = trainingSkill.TrainingSkill
		}
		if updatedTraining.TrainingCost != 0 {
			training.TrainingCost = updatedTraining.TrainingCost
		}
		if updatedTraining.EmployeeID != 0 {
			// Fetch employee data based on EmployeeID
			var employee models.Employee
			result = db.First(&employee, "id = ?", updatedTraining.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			training.EmployeeID = updatedTraining.EmployeeID
			training.FullNameEmployee = employee.FullName
		}
		if updatedTraining.StartDate != "" {
			training.StartDate = updatedTraining.StartDate
		}
		if updatedTraining.EndDate != "" {
			training.EndDate = updatedTraining.EndDate
		}
		if updatedTraining.Status != "" {
			training.Status = updatedTraining.Status
		}
		if updatedTraining.Description != "" {
			training.Description = updatedTraining.Description
		}

		// Update the training in the database
		if err := db.Save(&training).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update training"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Training updated successfully",
			"data":    training,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// DeleteTrainingByID memungkinkan hanya admin untuk menghapus data pelatihan berdasarkan ID.
func DeleteTrainingByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Verify if the user is an admin
		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Extract training ID from request parameters
		trainingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid training ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch training data by ID
		var training models.Training
		result = db.First(&training, trainingID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Training not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the training from the database
		if err := db.Delete(&training).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete training"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Provide success response
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Training deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
