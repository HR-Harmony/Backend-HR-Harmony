package controllers

import (
	"errors"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"log"
	"net/http"
	"strings"
	"time"
)

// CreateEmployeeAccountByAdmin handles the creation of an employee account by admin
func CreateClientAccountByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Bind the employee data from the request body
		var employee models.Employee
		if err := c.Bind(&employee); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate all employee data
		if employee.FirstName == "" || employee.LastName == "" || employee.ContactNumber == "" ||
			employee.Gender == "" || employee.Email == "" || employee.Username == "" ||
			employee.Password == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid client data. All fields are required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		passwordWithNoHash := employee.Password

		// Check if username is unique
		var existingUsername models.Employee
		result = db.Where("username = ?", employee.Username).First(&existingUsername)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Username already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to check username"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Check if contact number is unique
		var existingContactNumber models.Employee
		result = db.Where("contact_number = ?", employee.ContactNumber).First(&existingContactNumber)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Contact number already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to check contact number"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Check if email is unique
		var existingEmail models.Employee
		result = db.Where("email = ?", employee.Email).First(&existingEmail)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Email already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to check email"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Set the created timestamp
		currentTime := time.Now()
		employee.CreatedAt = &currentTime
		// Hash the password before saving to the database
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(employee.Password), bcrypt.DefaultCost)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to hash password"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}
		employee.Password = string(hashedPassword)

		employee.FullName = employee.FirstName + " " + employee.LastName

		employee.IsClient = true

		employee.IsActive = true

		// Create the employee account in the database
		db.Create(&employee)

		// Send email notification to the employee with the plain text password
		err = helper.SendClientAccountNotificationWithPlainTextPassword(employee.Email, employee.FirstName+" "+employee.LastName, employee.Username, passwordWithNoHash)
		if err != nil {
			log.Println("Failed to send welcome email:", err)
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send welcome email"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Client account created successfully",
			"employee": map[string]interface{}{
				"id":             employee.ID,
				"first_name":     employee.FirstName,
				"last_name":      employee.LastName,
				"full_name":      employee.FullName,
				"contact_number": employee.ContactNumber,
				"gender":         employee.Gender,
				"email":          employee.Email,
				"username":       employee.Username,
				"country":        employee.Country,
				"is_active":      employee.IsActive,
			},
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllClientsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Check if the user is an admin
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

		// Fetch client employees from the database
		var clientEmployees []struct {
			ID            uint   `json:"id"`
			FirstName     string `json:"first_name"`
			LastName      string `json:"last_name"`
			FullName      string `json:"full_name"`
			ContactNumber string `json:"contact_number"`
			Gender        string `json:"gender"`
			Email         string `json:"email"`
			Username      string `json:"username"`
			Country       string `json:"country"`
			IsActive      bool   `json:"is_active"`
		}
		db.Model(&models.Employee{}).Where("is_client = ?", true).
			Select("id", "first_name", "last_name", "full_name", "contact_number", "gender", "email", "username", "country", "is_active").
			Find(&clientEmployees)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Client  data retrieved successfully",
			"data":    clientEmployees,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetClientByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		employeeID := c.Param("id")
		if employeeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Client ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var employee models.Employee
		result = db.First(&employee, "id = ? AND is_client = ?", employeeID, true)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Client not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		employeeResponse := map[string]interface{}{
			"id":             employee.ID,
			"first_name":     employee.FirstName,
			"last_name":      employee.LastName,
			"full_name":      employee.FullName,
			"contact_number": employee.ContactNumber,
			"gender":         employee.Gender,
			"email":          employee.Email,
			"username":       employee.Username,
			"country":        employee.Country,
			"is_active":      employee.IsActive,
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"message":  "Client retrieved successfully",
			"employee": employeeResponse,
		})
	}
}

func UpdateClientAccountByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"})
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"})
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"})
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			return c.JSON(http.StatusNotFound, helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"})
		}

		if !adminUser.IsAdminHR {
			return c.JSON(http.StatusForbidden, helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"})
		}

		employeeID := c.Param("id")
		if employeeID == "" {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Client ID is missing"})
		}

		var existingEmployee models.Employee
		result = db.First(&existingEmployee, "id = ?", employeeID)
		if result.Error != nil {
			return c.JSON(http.StatusNotFound, helper.ErrorResponse{Code: http.StatusNotFound, Message: "Client not found"})
		}

		var updatedEmployee models.Employee
		if err := c.Bind(&updatedEmployee); err != nil {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"})
		}

		if updatedEmployee.FirstName != "" {
			existingEmployee.FirstName = updatedEmployee.FirstName
			existingEmployee.FullName = existingEmployee.FirstName + " " + existingEmployee.LastName // Update full name
		}
		if updatedEmployee.LastName != "" {
			existingEmployee.LastName = updatedEmployee.LastName
			existingEmployee.FullName = existingEmployee.FirstName + " " + existingEmployee.LastName // Update full name
		}

		if updatedEmployee.ContactNumber != "" {
			existingEmployee.ContactNumber = updatedEmployee.ContactNumber
		}
		if updatedEmployee.Gender != "" {
			existingEmployee.Gender = updatedEmployee.Gender
		}
		if updatedEmployee.Email != "" {
			existingEmployee.Email = updatedEmployee.Email
		}

		if updatedEmployee.Country != "" {
			existingEmployee.Country = updatedEmployee.Country
		}

		if updatedEmployee.IsActive {
			existingEmployee.IsActive = updatedEmployee.IsActive
		} else {
			existingEmployee.IsActive = false
		}

		if updatedEmployee.Username != "" {
			existingEmployee.Username = updatedEmployee.Username
		}
		if updatedEmployee.Password != "" {
			// Hash the updated password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatedEmployee.Password), bcrypt.DefaultCost)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to hash password"})
			}
			existingEmployee.Password = string(hashedPassword)
		}

		updatedEmployee.FullName = updatedEmployee.FirstName + " " + updatedEmployee.LastName

		passwordWithNoHash := updatedEmployee.Password

		if err := db.Save(&existingEmployee).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update employee data"})
		}

		// Send email notification to the employee with the plain text password
		err = helper.SendEmployeeAccountNotificationWithPlainTextPassword(existingEmployee.Email, existingEmployee.FirstName+" "+existingEmployee.LastName, existingEmployee.Username, passwordWithNoHash)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send welcome email"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		employeeWithoutPayrollInfo := map[string]interface{}{
			"id":             existingEmployee.ID,
			"first_name":     existingEmployee.FirstName,
			"last_name":      existingEmployee.LastName,
			"full_name":      existingEmployee.FullName,
			"contact_number": existingEmployee.ContactNumber,
			"gender":         existingEmployee.Gender,
			"email":          existingEmployee.Email,
			"username":       existingEmployee.Username,
			"country":        existingEmployee.Country,
			"is_active":      existingEmployee.IsActive,
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"message":  "Client account updated successfully",
			"employee": employeeWithoutPayrollInfo,
		})
	}
}

func DeleteClientAccountByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"})
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"})
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"})
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			return c.JSON(http.StatusNotFound, helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"})
		}

		if !adminUser.IsAdminHR {
			return c.JSON(http.StatusForbidden, helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"})
		}

		employeeID := c.Param("id")
		if employeeID == "" {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Client ID is missing"})
		}

		var existingEmployee models.Employee
		result = db.First(&existingEmployee, "id = ?", employeeID)
		if result.Error != nil {
			return c.JSON(http.StatusNotFound, helper.ErrorResponse{Code: http.StatusNotFound, Message: "Client not found"})
		}

		// Hapus terlebih dahulu entri terkait di tabel payroll_infos
		if err := db.Where("employee_id = ?", existingEmployee.ID).Delete(&models.PayrollInfo{}).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete related payroll information"})
		}

		if err := db.Delete(&existingEmployee).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete employee"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Client deleted successfully",
		})
	}
}
