// controllers/createEmployeeAccount.go

package controllers

import (
	"errors"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"strings"
	"time"
)

// CreateEmployeeAccountByAdmin handles the creation of an employee account by admin
func CreateEmployeeAccountByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			employee.Password == "" || employee.ShiftID == 0 || employee.RoleID == 0 ||
			employee.DepartmentID == 0 || employee.BasicSalary == 0 || employee.HourlyRate == 0 ||
			employee.PaySlipType == "" || employee.DesignationID == 0 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid employee data. All fields are required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		passwordWithNoHash := employee.Password

		// Check if the department exists
		var officeShift models.Shift
		result = db.First(&officeShift, "id = ?", employee.ShiftID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid shift name. Shift not found."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		employee.Shift = officeShift.ShiftName

		// Check if the department exists
		var role models.Role
		result = db.First(&role, "id = ?", employee.RoleID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid role name. Role not found."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		employee.Role = role.RoleName

		// Check if the department exists
		var department models.Department
		result = db.First(&department, "id = ?", employee.DepartmentID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid department name. Department not found."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		employee.Department = department.DepartmentName

		// Check if the designation exists
		var designation models.Designation
		result = db.First(&designation, "id = ?", employee.DesignationID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid designation ID. Designation not found."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		employee.Designation = designation.DesignationName
		employee.DesignationID = designation.ID

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

		employee.ShiftID = officeShift.ID
		employee.RoleID = role.ID
		employee.DepartmentID = department.ID

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

		// Create the employee account in the database
		db.Create(&employee)

		// Send email notification to the employee with the plain text password
		err = helper.SendEmployeeAccountNotificationWithPlainTextPassword(employee.Email, employee.FirstName+" "+employee.LastName, employee.Username, passwordWithNoHash)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send welcome email"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Respond with success
		successResponse := helper.Response{
			Code:     http.StatusCreated,
			Error:    false,
			Message:  "Employee account created successfully",
			Employee: &employee,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// ExitEmployee handles the exit process for employees by admin
func ExitEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var exitData models.ExitEmployee
		if err := c.Bind(&exitData); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Parse exitDate from string to time.Time
		exitDate, err := time.Parse("2006-01-02", exitData.ExitDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid exitDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		exitData.ExitDate = exitDate.Format("2006-01-02")

		// Validasi apakah employee dengan ID yang diberikan ada
		var employee models.Employee
		result = db.First(&employee, exitData.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Validasi apakah exit dengan ID yang diberikan ada
		var exit models.Exit
		result = db.First(&exit, exitData.ExitID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Exit not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Update status IsActive berdasarkan disable account
		if !exitData.DisableAccount {
			employee.IsActive = true
		} else {
			employee.IsActive = false
		}

		// Membuat record ExitEmployee di database
		exitData.CreatedAt = time.Now()

		db.Create(&exitData)

		// Update data employee di database
		db.Save(&employee)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Employee exit processed successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetAllExitEmployees returns all ExitEmployee records for admin
func GetAllExitEmployees(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Fetch all ExitEmployee records
		var exitEmployees []models.ExitEmployee
		if err := db.Find(&exitEmployees).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch ExitEmployee records"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Respond with success
		successResponse := helper.Response{
			Code:          http.StatusOK,
			Error:         false,
			Message:       "ExitEmployee records retrieved successfully",
			ExitEmployees: exitEmployees,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetExitEmployeeByID returns the ExitEmployee record based on the provided ID for admin
func GetExitEmployeeByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Extract ExitEmployee ID from the request
		exitEmployeeID := c.Param("id")

		// Fetch the ExitEmployee record by ID
		var exitEmployee models.ExitEmployee
		if err := db.Where("id = ?", exitEmployeeID).First(&exitEmployee).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "ExitEmployee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with success
		successResponse := helper.Response{
			Code:         http.StatusOK,
			Error:        false,
			Message:      "ExitEmployee record retrieved successfully",
			ExitEmployee: &exitEmployee,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// DeleteExitEmployeeByID deletes the ExitEmployee record based on the provided ID for admin
func DeleteExitEmployeeByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Extract ExitEmployee ID from the request
		exitEmployeeID := c.Param("id")

		// Fetch the ExitEmployee record by ID
		var exitEmployee models.ExitEmployee
		if err := db.Where("id = ?", exitEmployeeID).First(&exitEmployee).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "ExitEmployee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the ExitEmployee record from the database
		if err := db.Delete(&exitEmployee).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to delete ExitEmployee record"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "ExitEmployee record deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
