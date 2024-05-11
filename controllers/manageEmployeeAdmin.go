// controllers/createEmployeeAccount.go

package controllers

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"log"
	"net/http"
	"strconv"
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

		payrollID := generateUniquePayrollID()
		employee.PayrollID = payrollID

		employee.FullName = employee.FirstName + " " + employee.LastName

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
			log.Println("Failed to send welcome email:", err)
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

// GetAllEmployeesByAdmin handles the retrieval of all employees by admin with pagination
func GetAllEmployeesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employees []models.Employee
		db.Where("is_client = ?", false).Offset(offset).Limit(perPage).Find(&employees)

		var employeesResponse []helper.EmployeeResponse
		for _, emp := range employees {
			employeeResponse := helper.EmployeeResponse{
				ID:            emp.ID,
				PayrollID:     emp.PayrollID,
				FirstName:     emp.FirstName,
				LastName:      emp.LastName,
				ContactNumber: emp.ContactNumber,
				Gender:        emp.Gender,
				Email:         emp.Email,
				Username:      emp.Username,
				Password:      emp.Password,
				ShiftID:       emp.ShiftID,
				Shift:         emp.Shift,
				RoleID:        emp.RoleID,
				Role:          emp.Role,
				DepartmentID:  emp.DepartmentID,
				Department:    emp.Department,
				DesignationID: emp.DesignationID,
				Designation:   emp.Designation,
				BasicSalary:   emp.BasicSalary,
				HourlyRate:    emp.HourlyRate,
				PaySlipType:   emp.PaySlipType,
				IsActive:      emp.IsActive,
				PaidStatus:    emp.PaidStatus,
				CreatedAt:     emp.CreatedAt,
				UpdatedAt:     emp.UpdatedAt,
			}
			employeesResponse = append(employeesResponse, employeeResponse)
		}

		var totalCount int64
		db.Model(&models.Employee{}).Where("is_client = ?", false).Count(&totalCount)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":      http.StatusOK,
			"error":     false,
			"message":   "All employees retrieved successfully",
			"employees": employeesResponse,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		})
	}
}

func GetEmployeeByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var employee models.Employee
		result = db.First(&employee, "id = ?", employeeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		employeeResponse := helper.EmployeeResponse{
			ID:            employee.ID,
			PayrollID:     employee.PayrollID,
			FirstName:     employee.FirstName,
			LastName:      employee.LastName,
			ContactNumber: employee.ContactNumber,
			Gender:        employee.Gender,
			Email:         employee.Email,
			Username:      employee.Username,
			Password:      employee.Password,
			ShiftID:       employee.ShiftID,
			Shift:         employee.Shift,
			RoleID:        employee.RoleID,
			Role:          employee.Role,
			DepartmentID:  employee.DepartmentID,
			Department:    employee.Department,
			DesignationID: employee.DesignationID,
			Designation:   employee.Designation,
			BasicSalary:   employee.BasicSalary,
			HourlyRate:    employee.HourlyRate,
			PaySlipType:   employee.PaySlipType,
			IsActive:      employee.IsActive,
			PaidStatus:    employee.PaidStatus,
			CreatedAt:     employee.CreatedAt,
			UpdatedAt:     employee.UpdatedAt,
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"message":  "Employee retrieved successfully",
			"employee": employeeResponse,
		})
	}
}

func UpdateEmployeeAccountByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee ID is missing"})
		}

		var existingEmployee models.Employee
		result = db.First(&existingEmployee, "id = ?", employeeID)
		if result.Error != nil {
			return c.JSON(http.StatusNotFound, helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"})
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

		passwordWithNoHash := updatedEmployee.Password

		if updatedEmployee.ShiftID != 0 {
			var officeShift models.Shift
			result = db.First(&officeShift, "id = ?", updatedEmployee.ShiftID)
			if result.Error != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid shift name. Shift not found."})
			}
			existingEmployee.ShiftID = updatedEmployee.ShiftID
			existingEmployee.Shift = officeShift.ShiftName
		}
		if updatedEmployee.RoleID != 0 {
			var role models.Role
			result = db.First(&role, "id = ?", updatedEmployee.RoleID)
			if result.Error != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid role name. Role not found."})
			}
			existingEmployee.RoleID = updatedEmployee.RoleID
			existingEmployee.Role = role.RoleName
		}
		if updatedEmployee.DepartmentID != 0 {
			var department models.Department
			result = db.First(&department, "id = ?", updatedEmployee.DepartmentID)
			if result.Error != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid department name. Department not found."})
			}
			existingEmployee.DepartmentID = updatedEmployee.DepartmentID
			existingEmployee.Department = department.DepartmentName
		}
		if updatedEmployee.DesignationID != 0 {
			var designation models.Designation
			result = db.First(&designation, "id = ?", updatedEmployee.DesignationID)
			if result.Error != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid designation ID. Designation not found."})
			}
			existingEmployee.DesignationID = updatedEmployee.DesignationID
			existingEmployee.Designation = designation.DesignationName
		}
		if updatedEmployee.BasicSalary != 0 {
			existingEmployee.BasicSalary = updatedEmployee.BasicSalary
		}
		if updatedEmployee.HourlyRate != 0 {
			existingEmployee.HourlyRate = updatedEmployee.HourlyRate
		}
		if updatedEmployee.PaySlipType != "" {
			existingEmployee.PaySlipType = updatedEmployee.PaySlipType
		}

		if err := db.Save(&existingEmployee).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update employee data"})
		}

		// Send email notification to the employee with the plain text password
		err = helper.SendEmployeeAccountNotificationWithPlainTextPassword(existingEmployee.Email, existingEmployee.FirstName+" "+existingEmployee.LastName, existingEmployee.Username, passwordWithNoHash)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send welcome email"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Exclude PayrollInfo from the response
		employeeWithoutPayrollInfo := helper.EmployeeResponse{
			ID:            existingEmployee.ID,
			PayrollID:     existingEmployee.PayrollID,
			FirstName:     existingEmployee.FirstName,
			LastName:      existingEmployee.LastName,
			ContactNumber: existingEmployee.ContactNumber,
			Gender:        existingEmployee.Gender,
			Email:         existingEmployee.Email,
			Username:      existingEmployee.Username,
			Password:      existingEmployee.Password,
			ShiftID:       existingEmployee.ShiftID,
			Shift:         existingEmployee.Shift,
			RoleID:        existingEmployee.RoleID,
			Role:          existingEmployee.Role,
			DepartmentID:  existingEmployee.DepartmentID,
			Department:    existingEmployee.Department,
			DesignationID: existingEmployee.DesignationID,
			Designation:   existingEmployee.Designation,
			BasicSalary:   existingEmployee.BasicSalary,
			HourlyRate:    existingEmployee.HourlyRate,
			PaySlipType:   existingEmployee.PaySlipType,
			IsActive:      existingEmployee.IsActive,
			PaidStatus:    existingEmployee.PaidStatus,
			CreatedAt:     existingEmployee.CreatedAt,
			UpdatedAt:     existingEmployee.UpdatedAt,
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"message":  "Employee account updated successfully",
			"employee": employeeWithoutPayrollInfo,
		})
	}
}

func DeleteEmployeeAccountByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee ID is missing"})
		}

		var existingEmployee models.Employee
		result = db.First(&existingEmployee, "id = ?", employeeID)
		if result.Error != nil {
			return c.JSON(http.StatusNotFound, helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"})
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
			"message": "Employee deleted successfully",
		})
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

		// Check if exit record already exists for the employee
		var existingExit models.ExitEmployee
		result = db.Where("employee_id = ?", exitData.EmployeeID).First(&existingExit)
		if result.RowsAffected > 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Employee already exited"}
			return c.JSON(http.StatusConflict, errorResponse)
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

// GetAllExitEmployees returns all ExitEmployee records for admin with pagination
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

		// Fetch all ExitEmployee records with pagination
		var exitEmployees []models.ExitEmployee
		if err := db.Offset(offset).Limit(perPage).Find(&exitEmployees).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch ExitEmployee records"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Get total count of ExitEmployee records
		var totalCount int64
		db.Model(&models.ExitEmployee{}).Count(&totalCount)

		// Respond with success
		successResponse := map[string]interface{}{
			"Code":          http.StatusOK,
			"Error":         false,
			"Message":       "ExitEmployee records retrieved successfully",
			"ExitEmployees": exitEmployees,
			"Pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
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

// UpdateEmployeePasswordByAdmin handles updating an employee's password by admin
func UpdateEmployeePasswordByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Get employee ID from request parameters
		employeeID := c.Param("id")
		if employeeID == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Employee ID is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Convert employee ID to uint
		employeeIDUint, err := strconv.ParseUint(employeeID, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid employee ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Bind the new password and repeat password from the request body
		var newPassword struct {
			NewPassword    string `json:"new_password"`
			RepeatPassword string `json:"repeat_password"`
		}
		if err := c.Bind(&newPassword); err != nil {
			// Handle invalid request body
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate the new password and repeat password
		if newPassword.NewPassword == "" || newPassword.RepeatPassword == "" {
			// Handle missing new password or repeat password
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "New password and repeat password are required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if newPassword.NewPassword != newPassword.RepeatPassword {
			// Handle mismatch between new password and repeat password
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "New password and repeat password do not match"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the employee from the database
		var employee models.Employee
		result = db.First(&employee, "id = ?", employeeIDUint)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			} else {
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to hash password"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Update employee's password
		employee.Password = string(hashedPassword)
		db.Save(&employee)

		// Send password change notification to the employee
		go func(email, fullName, newPassword string) {
			if err := helper.SendPasswordChangeNotification(email, fullName, newPassword); err != nil {
				fmt.Println("Failed to send password change notification email:", err)
			}
		}(employee.Email, employee.FirstName+" "+employee.LastName, newPassword.NewPassword)

		// Respond with success
		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Employee password updated successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func generateUniquePayrollID() int64 {
	// Generate a new UUID
	uid, err := uuid.NewUUID()
	if err != nil {
		panic(err.Error())
	}

	// Get the 64-bit unsigned integer representation of the UUID
	uidInt := uid.ID()

	// Convert the unsigned integer to a signed int64
	uidInt64 := int64(uidInt)

	return uidInt64
}
