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
			employee.Password == "" || employee.Shift == "" || employee.Role == "" ||
			employee.Department == "" || employee.BasicSalary == 0 || employee.HourlyRate == 0 ||
			employee.PaySlipType == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid employee data. All fields are required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		passwordWithNoHash := employee.Password

		// Check if the department exists
		var officeShift models.Shift
		result = db.First(&officeShift, "shift_name = ?", employee.Shift)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid shift name. Shift not found."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the department exists
		var role models.Role
		result = db.First(&role, "role_name = ?", employee.Role)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid role name. Role not found."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the department exists
		var department models.Department
		result = db.First(&department, "department_name = ?", employee.Department)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid department name. Department not found."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

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
