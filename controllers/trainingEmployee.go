package controllers

import (
	"errors"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"strconv"
	"strings"
)

// GetTrainingByEmployeeID mengambil data pelatihan milik karyawan berdasarkan ID karyawan
func GetTrainingByEmployeeID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Retrieve training data for the employee
		var trainings []models.Training
		db.Where("employee_id = ?", employee.ID).Find(&trainings)

		// Return the training data
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Training data retrieved successfully",
			"data":    trainings,
		})
	}
}

// GetTrainingByIDAndEmployeeID mengambil data training milik karyawan berdasarkan ID training dan ID karyawan
func GetTrainingByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Extract trainingID from the request parameters
		trainingID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid training ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve training data for the specified ID
		var training models.Training
		result = db.Where("id = ?", trainingID).First(&training)
		if result.Error != nil {
			// Check if the training ID is not found
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Training ID not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch training data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Check if the retrieved training belongs to the employee
		if training.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Training does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Return the training data
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Training data retrieved successfully",
			"data":    training,
		})
	}
}
