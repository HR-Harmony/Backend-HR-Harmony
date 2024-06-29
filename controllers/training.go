package controllers

import (
	"fmt"
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

func CreateTrainerByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var trainer models.Trainer
		if err := c.Bind(&trainer); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(trainer.FirstName) < 1 || len(trainer.FirstName) > 100 || !regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(trainer.FirstName) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "First name must be between 1 and 100 characters and contain only letters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(trainer.LastName) < 1 || len(trainer.LastName) > 100 || !regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(trainer.FirstName) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Last name must be between 1 and 100 characters and contain only letters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(trainer.Email) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid email format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		contactNumberRegex := regexp.MustCompile(`^\d{10,14}$`)
		if !contactNumberRegex.MatchString(trainer.ContactNumber) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Contact number must be between 10 and 14 digits and contain only numbers"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(trainer.Expertise) < 5 || len(trainer.Expertise) > 100 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Expertise must be between 1 and 100 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(trainer.Address) < 5 || len(trainer.Address) > 1000 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Address must be between 1 and 1000 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		trainer.FullName = trainer.FirstName + " " + trainer.LastName

		if err := db.Create(&trainer).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create trainer"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

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

		searching := c.QueryParam("searching")

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

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
		query.Order("id DESC").Offset(offset).Limit(perPage).Find(&trainers)

		var totalCount int64
		db.Model(&models.Trainer{}).Count(&totalCount)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Trainers fetched successfully",
			"data":    trainers,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetTrainerByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		trainerID := c.Param("id")
		if trainerID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Trainer ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var trainer models.Trainer
		result = db.First(&trainer, "id = ?", trainerID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Trainer not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

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

		trainerID := c.Param("id")
		if trainerID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Trainer ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var trainer models.Trainer
		result = db.First(&trainer, "id = ?", trainerID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Trainer not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedTrainer models.Trainer
		if err := c.Bind(&updatedTrainer); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		/*
			if updatedTrainer.FirstName != "" {
				trainer.FirstName = updatedTrainer.FirstName
				trainer.FullName = trainer.FirstName + " " + trainer.LastName // Update full name
			}
		*/

		if updatedTrainer.FirstName != "" {
			if len(updatedTrainer.FirstName) < 1 || len(updatedTrainer.FirstName) > 100 || !regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(updatedTrainer.FirstName) {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "First Name must be between 1 and 100 and contain only letters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			trainer.FirstName = updatedTrainer.FirstName
			trainer.FullName = trainer.FirstName + " " + trainer.LastName // Update full name
		}

		/*
			if updatedTrainer.LastName != "" {
				trainer.LastName = updatedTrainer.LastName
				trainer.FullName = trainer.FirstName + " " + trainer.LastName // Update full name
			}
		*/

		if updatedTrainer.LastName != "" {
			if len(updatedTrainer.LastName) < 1 || len(updatedTrainer.LastName) > 100 || !regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(updatedTrainer.LastName) {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Last Name must be between 1 and 100 and contain only letters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			trainer.LastName = updatedTrainer.LastName
			trainer.FullName = trainer.FirstName + " " + trainer.LastName // Update full name
		}

		/*
			if updatedTrainer.ContactNumber != "" {
				trainer.ContactNumber = updatedTrainer.ContactNumber
			}
		*/

		if updatedTrainer.ContactNumber != "" {
			contactNumberRegex := regexp.MustCompile(`^\d{10,14}$`)
			if !contactNumberRegex.MatchString(updatedTrainer.ContactNumber) {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Contact number must be between 10 and 14 digits and contain only numbers"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			trainer.ContactNumber = updatedTrainer.ContactNumber
		}

		/*
			if updatedTrainer.Email != "" {
				trainer.Email = updatedTrainer.Email
			}
		*/

		if updatedTrainer.Email != "" {
			emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
			if !emailRegex.MatchString(updatedTrainer.Email) {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid email format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			trainer.Email = updatedTrainer.Email
		}

		/*
			if updatedTrainer.Expertise != "" {
				trainer.Expertise = updatedTrainer.Expertise
			}
		*/

		if updatedTrainer.Expertise != "" {
			if len(updatedTrainer.Expertise) < 5 || len(updatedTrainer.Expertise) > 100 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Expertise must be between 5 and 100 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			trainer.Expertise = updatedTrainer.Expertise
		}

		/*
			if updatedTrainer.Address != "" {
				trainer.Address = updatedTrainer.Address
			}
		*/

		if updatedTrainer.Address != "" {
			if len(updatedTrainer.Address) < 5 || len(updatedTrainer.Address) > 1000 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Address must be between 5 and 1000 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			trainer.Address = updatedTrainer.Address
		}

		db.Save(&trainer)

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

		trainerID := c.Param("id")
		if trainerID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Trainer ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var trainer models.Trainer
		result = db.First(&trainer, "id = ?", trainerID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Trainer not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&trainer)

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

		var trainingSkill models.TrainingSkill
		if err := c.Bind(&trainingSkill); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if err := db.Create(&trainingSkill).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create TrainingSkill"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

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

		searching := strings.ToLower(c.QueryParam("searching"))

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		var trainingSkills []models.TrainingSkill
		query := db.Model(&trainingSkills)
		if searching != "" {
			query = query.Where("LOWER(training_skill) LIKE ?", "%"+searching+"%")
		}
		query.Order("id DESC").Offset(offset).Limit(perPage).Find(&trainingSkills)

		var totalCount int64
		db.Model(&models.TrainingSkill{}).Count(&totalCount)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "TrainingSkills fetched successfully",
			"data":    trainingSkills,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetTrainingSkillByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		id := c.Param("id")

		var trainingSkill models.TrainingSkill
		if err := db.First(&trainingSkill, "id = ?", id).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "TrainingSkill not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

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

		id := c.Param("id")

		var updatedTrainingSkill models.TrainingSkill
		if err := c.Bind(&updatedTrainingSkill); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var trainingSkill models.TrainingSkill
		if err := db.First(&trainingSkill, "id = ?", id).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "TrainingSkill not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if err := db.Model(&trainingSkill).Updates(updatedTrainingSkill).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update TrainingSkill"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

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

		id := c.Param("id")

		var trainingSkill models.TrainingSkill
		if err := db.First(&trainingSkill, "id = ?", id).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "TrainingSkill not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if err := db.Delete(&trainingSkill).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete TrainingSkill"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

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

		var training models.Training
		if err := c.Bind(&training); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		_, err = time.Parse("2006-01-02", training.StartDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid start_date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		_, err = time.Parse("2006-01-02", training.EndDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid end_date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var trainer models.Trainer
		result = db.First(&trainer, "id = ?", training.TrainerID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Trainer not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		training.FullNameTrainer = trainer.FullName

		var trainingSkill models.TrainingSkill
		result = db.First(&trainingSkill, "id = ?", training.TrainingSkillID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Training skill not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		training.TrainingSkill = trainingSkill.TrainingSkill

		var employee models.Employee
		result = db.First(&employee, "id = ?", training.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		training.FullNameEmployee = employee.FullName

		training.Status = "Pending"

		if err := db.Create(&training).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create training"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Mengirim notifikasi email kepada karyawan
		err = helper.SendTrainingNotification(employee.Email, employee.FullName, trainer.FullName, trainingSkill.TrainingSkill, training.StartDate, training.EndDate)
		if err != nil {
			fmt.Println("Failed to send training notification email:", err)
			// Tangani kesalahan sesuai kebutuhan Anda, misalnya dengan memberikan respons ke klien atau mencatat log
		}

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

		searching := c.QueryParam("searching")

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		var trainings []models.Training
		query := db.Model(&trainings)
		if searching != "" {
			searching = strings.ToLower(searching)
			query = query.Where("LOWER(full_name_trainer) LIKE ? OR LOWER(training_skill) LIKE ? OR LOWER(full_name_employee) LIKE ? OR CAST(training_cost AS VARCHAR) LIKE ?",
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
				searching,
			)
		}
		query.Order("id DESC").Offset(offset).Limit(perPage).Find(&trainings)

		var totalCount int64
		db.Model(&models.Training{}).Count(&totalCount)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Trainings fetched successfully",
			"data":    trainings,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetTrainingByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		trainingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid training ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var training models.Training
		result = db.First(&training, trainingID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Training not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Training fetched successfully",
			"data":    training,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateTrainingByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		trainingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid training ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var training models.Training
		result = db.First(&training, trainingID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Training not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedTraining models.Training
		if err := c.Bind(&updatedTraining); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedTraining.TrainerID != 0 {
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
			var employee models.Employee
			result = db.First(&employee, "id = ?", updatedTraining.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			training.EmployeeID = updatedTraining.EmployeeID
			training.FullNameEmployee = employee.FullName
		}

		if updatedTraining.GoalTypeID != 0 {
			var goalType models.GoalType
			result = db.First(&goalType, "id = ?", updatedTraining.GoalTypeID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid goal type ID. Goal type not found."}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			updatedTraining.GoalTypeID = goalType.ID
			updatedTraining.GoalType = goalType.GoalType
			training.GoalTypeID = updatedTraining.GoalTypeID
			training.GoalType = updatedTraining.GoalType
		}

		if updatedTraining.Performance != "" {
			training.Performance = updatedTraining.Performance
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

		if err := db.Save(&training).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update training"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Training updated successfully",
			"data":    training,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteTrainingByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		trainingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid training ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var training models.Training
		result = db.First(&training, trainingID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Training not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if err := db.Delete(&training).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete training"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Training deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
