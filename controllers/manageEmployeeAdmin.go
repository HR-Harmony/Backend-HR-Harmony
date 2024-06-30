// controllers/createEmployeeAccount.go

package controllers

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CreateMultipleEmployeeAccountsByAdmin handles the creation of multiple employee accounts by admin from an Excel file
func CreateMultipleEmployeeAccountsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Parse the uploaded Excel file
		file, err := c.FormFile("file")
		if err != nil {
			log.Println("Invalid file:", err)
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid file"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		src, err := file.Open()
		if err != nil {
			log.Println("Failed to open file:", err)
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to open file"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}
		defer src.Close()

		excelFile, err := excelize.OpenReader(src)
		if err != nil {
			log.Println("Failed to read Excel file:", err)
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to read Excel file"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		sheetName := excelFile.GetSheetName(0)
		rows, err := excelFile.GetRows(sheetName)
		if err != nil {
			log.Println("Failed to read Excel rows:", err)
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to read Excel rows"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var createdEmployees []models.Employee
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				log.Println("Transaction rollback due to panic:", r)
			}
		}()

		for _, row := range rows[1:] { // Skipping header row
			if len(row) < 14 {
				log.Println("Skipping incomplete row:", row)
				continue // Skip incomplete rows
			}

			shiftID, err := strconv.ParseUint(row[7], 10, 32)
			if err != nil {
				log.Println("Invalid shift ID:", row[7])
				continue // Skip invalid shift ID
			}

			roleID, err := strconv.ParseUint(row[8], 10, 32)
			if err != nil {
				log.Println("Invalid role ID:", row[8])
				continue // Skip invalid role ID
			}

			designationID, err := strconv.ParseUint(row[9], 10, 32)
			if err != nil {
				log.Println("Invalid designation ID:", row[9])
				continue // Skip invalid designation ID
			}

			departmentID, err := strconv.ParseUint(row[10], 10, 32)
			if err != nil {
				log.Println("Invalid department ID:", row[10])
				continue // Skip invalid department ID
			}

			basicSalary, err := strconv.ParseFloat(row[11], 64)
			if err != nil {
				log.Println("Invalid basic salary:", row[11])
				continue // Skip invalid basic salary
			}

			hourlyRate, err := strconv.ParseFloat(row[12], 64)
			if err != nil {
				log.Println("Invalid hourly rate:", row[12])
				continue // Skip invalid hourly rate
			}

			employee := models.Employee{
				FirstName:     row[0],
				LastName:      row[1],
				ContactNumber: row[2],
				Gender:        row[3],
				Email:         row[4],
				Username:      row[5],
				Password:      row[6],
				ShiftID:       uint(shiftID),
				RoleID:        uint(roleID),
				DesignationID: uint(designationID),
				DepartmentID:  uint(departmentID),
				BasicSalary:   basicSalary,
				HourlyRate:    hourlyRate,
				PaySlipType:   row[13],
			}

			passwordWithNoHash := employee.Password

			// Validate all employee data
			if employee.FirstName == "" || employee.LastName == "" || employee.ContactNumber == "" ||
				employee.Gender == "" || employee.Email == "" || employee.Username == "" ||
				employee.Password == "" || employee.ShiftID == 0 || employee.RoleID == 0 ||
				employee.DepartmentID == 0 || employee.BasicSalary == 0 || employee.HourlyRate == 0 ||
				employee.PaySlipType == "" || employee.DesignationID == 0 {
				log.Println("Skipping invalid data:", employee)
				continue // Skip invalid data
			}

			// Check if the shift exists
			var officeShift models.Shift
			result := tx.First(&officeShift, "id = ?", employee.ShiftID)
			if result.Error != nil {
				log.Println("Skipping invalid shift ID:", employee.ShiftID)
				continue // Skip invalid shift
			}

			// Check if the role exists
			var role models.Role
			result = tx.First(&role, "id = ?", employee.RoleID)
			if result.Error != nil {
				log.Println("Skipping invalid role ID:", employee.RoleID)
				continue // Skip invalid role
			}

			// Check if the department exists
			var department models.Department
			result = tx.First(&department, "id = ?", employee.DepartmentID)
			if result.Error != nil {
				log.Println("Skipping invalid department ID:", employee.DepartmentID)
				continue // Skip invalid department
			}

			// Check if the designation exists
			var designation models.Designation
			result = tx.First(&designation, "id = ?", employee.DesignationID)
			if result.Error != nil {
				log.Println("Skipping invalid designation ID:", employee.DesignationID)
				continue // Skip invalid designation
			}
			employee.DesignationID = designation.ID

			// Check if username is unique
			var existingUsername models.Employee
			result = tx.Where("username = ?", employee.Username).First(&existingUsername)
			if result.Error == nil {
				log.Println("Skipping duplicate username:", employee.Username)
				continue // Skip duplicate username
			} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				log.Println("Error checking username uniqueness:", result.Error)
				continue // Skip on error
			}

			// Check if contact number is unique
			var existingContactNumber models.Employee
			result = tx.Where("contact_number = ?", employee.ContactNumber).First(&existingContactNumber)
			if result.Error == nil {
				log.Println("Skipping duplicate contact number:", employee.ContactNumber)
				continue // Skip duplicate contact number
			} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				log.Println("Error checking contact number uniqueness:", result.Error)
				continue // Skip on error
			}

			// Check if email is unique
			var existingEmail models.Employee
			result = tx.Where("email = ?", employee.Email).First(&existingEmail)
			if result.Error == nil {
				log.Println("Skipping duplicate email:", employee.Email)
				continue // Skip duplicate email
			} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				log.Println("Error checking email uniqueness:", result.Error)
				continue // Skip on error
			}

			payrollID := generateUniquePayrollID()
			employee.PayrollID = payrollID

			employee.FullName = employee.FirstName + " " + employee.LastName

			// Hash the password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(employee.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Println("Failed to hash password:", err)
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to hash password"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
			employee.Password = string(hashedPassword)

			if err := tx.Create(&employee).Error; err != nil {
				log.Println("Failed to create employee:", err)
				tx.Rollback()
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to create employee"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}

			createdEmployees = append(createdEmployees, employee)

			go func() {
				err = helper.SendEmployeeAccountNotificationWithPlainTextPassword(employee.Email, employee.FullName, employee.Username, passwordWithNoHash)
				if err != nil {
					log.Println("Failed to send email:", err)
				}
			}()
		}

		if err := tx.Commit().Error; err != nil {
			log.Println("Failed to commit transaction:", err)
			tx.Rollback()
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to commit transaction"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		log.Println("Successfully created employees:", createdEmployees)
		successResponse := map[string]interface{}{"Code": http.StatusOK, "Error": false, "Message": "Employee accounts created successfully", "Data": createdEmployees}
		return c.JSON(http.StatusOK, successResponse)
	}
}

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

		if len(employee.FirstName) < 1 || len(employee.FirstName) > 30 || !regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(employee.FirstName) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "First Name must be between 1 and 30 characters and contain only letters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		/*
			if len(employee.LastName) < 1 || len(employee.LastName) > 30 || !regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(employee.LastName) {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Last Name must be between 1 and 30 characters and contain only letters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		*/

		if len(employee.Username) < 5 || len(employee.Username) > 15 || !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(employee.Username) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Username must be between 5 and 15 characters and contain only letters and numbers"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate Email using regexp
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(employee.Email) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid email format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate ContactNumber using regexp
		contactNumberRegex := regexp.MustCompile(`^\d{10,14}$`)
		if !contactNumberRegex.MatchString(employee.ContactNumber) {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Contact number must be between 10 and 14 digits and contain only numbers"}
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

		// Check if the department exists
		var role models.Role
		result = db.First(&role, "id = ?", employee.RoleID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid role name. Role not found."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the department exists
		var department models.Department
		result = db.First(&department, "id = ?", employee.DepartmentID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid department name. Department not found."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the designation exists
		var designation models.Designation
		result = db.First(&designation, "id = ?", employee.DesignationID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid designation ID. Designation not found."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

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

		if len(employee.Username) < 6 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Username must be more than 5 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

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

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		searching := c.QueryParam("searching")

		var employees []models.Employee
		query := db.Preload("Shift").Preload("Role").Preload("Department").Preload("Designation").
			Where("is_client = ? AND is_exit = ?", false, false).
			Order("id DESC").
			Offset(offset).
			Limit(perPage)

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where(
				db.Where("full_name ILIKE ?", searchPattern).
					Or("designation ILIKE ?", searchPattern).
					Or("contact_number ILIKE ?", searchPattern).
					Or("gender ILIKE ?", searchPattern).
					Or("country ILIKE ?", searchPattern).
					Or("role ILIKE ?", searchPattern))
		}

		if err := query.Find(&employees).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Error fetching employees"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var employeesResponse []helper.EmployeeResponse
		for _, emp := range employees {
			employeeResponse := helper.EmployeeResponse{
				ID:                       emp.ID,
				PayrollID:                emp.PayrollID,
				FirstName:                emp.FirstName,
				LastName:                 emp.LastName,
				FullName:                 emp.FullName,
				ContactNumber:            emp.ContactNumber,
				Gender:                   emp.Gender,
				Email:                    emp.Email,
				BirthdayDate:             emp.BirthdayDate,
				Username:                 emp.Username,
				Password:                 emp.Password,
				ShiftID:                  emp.ShiftID,
				Shift:                    emp.Shift.ShiftName,
				RoleID:                   emp.RoleID,
				Role:                     emp.Role.RoleName,
				DepartmentID:             emp.DepartmentID,
				Department:               emp.Department.DepartmentName,
				DesignationID:            emp.DesignationID,
				Designation:              emp.Designation.DesignationName,
				BasicSalary:              emp.BasicSalary,
				HourlyRate:               emp.HourlyRate,
				PaySlipType:              emp.PaySlipType,
				IsActive:                 emp.IsActive,
				PaidStatus:               emp.PaidStatus,
				MaritalStatus:            emp.MaritalStatus,
				Religion:                 emp.Religion,
				BloodGroup:               emp.BloodGroup,
				Nationality:              emp.Nationality,
				Citizenship:              emp.Citizenship,
				BpjsKesehatan:            emp.BpjsKesehatan,
				Address1:                 emp.Address1,
				Address2:                 emp.Address2,
				City:                     emp.City,
				StateProvince:            emp.StateProvince,
				ZipPostalCode:            emp.ZipPostalCode,
				Bio:                      emp.Bio,
				FacebookURL:              emp.FacebookURL,
				InstagramURL:             emp.InstagramURL,
				TwitterURL:               emp.TwitterURL,
				LinkedinURL:              emp.LinkedinURL,
				AccountTitle:             emp.AccountTitle,
				AccountNumber:            emp.AccountNumber,
				BankName:                 emp.BankName,
				Iban:                     emp.Iban,
				SwiftCode:                emp.SwiftCode,
				BankBranch:               emp.BankBranch,
				EmergencyContactFullName: emp.EmergencyContactFullName,
				EmergencyContactNumber:   emp.EmergencyContactNumber,
				EmergencyContactEmail:    emp.EmergencyContactEmail,
				EmergencyContactAddress:  emp.EmergencyContactAddress,
				CreatedAt:                emp.CreatedAt,
				UpdatedAt:                emp.UpdatedAt,
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
		result = db.Preload("Shift").Preload("Role").Preload("Department").Preload("Designation").First(&employee, "id = ?", employeeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		employeeResponse := helper.EmployeeResponse{
			ID:                       employee.ID,
			PayrollID:                employee.PayrollID,
			FirstName:                employee.FirstName,
			LastName:                 employee.LastName,
			FullName:                 employee.FullName,
			ContactNumber:            employee.ContactNumber,
			Gender:                   employee.Gender,
			Email:                    employee.Email,
			Username:                 employee.Username,
			Password:                 employee.Password,
			ShiftID:                  employee.ShiftID,
			Shift:                    employee.Shift.ShiftName,
			RoleID:                   employee.RoleID,
			Role:                     employee.Role.RoleName,
			DepartmentID:             employee.DepartmentID,
			Department:               employee.Department.DepartmentName,
			DesignationID:            employee.DesignationID,
			Designation:              employee.Designation.DesignationName,
			BasicSalary:              employee.BasicSalary,
			HourlyRate:               employee.HourlyRate,
			PaySlipType:              employee.PaySlipType,
			IsActive:                 employee.IsActive,
			PaidStatus:               employee.PaidStatus,
			MaritalStatus:            employee.MaritalStatus,
			Religion:                 employee.Religion,
			BloodGroup:               employee.BloodGroup,
			Nationality:              employee.Nationality,
			Citizenship:              employee.Citizenship,
			BpjsKesehatan:            employee.BpjsKesehatan,
			Address1:                 employee.Address1,
			Address2:                 employee.Address2,
			City:                     employee.City,
			StateProvince:            employee.StateProvince,
			ZipPostalCode:            employee.ZipPostalCode,
			Bio:                      employee.Bio,
			FacebookURL:              employee.FacebookURL,
			InstagramURL:             employee.InstagramURL,
			TwitterURL:               employee.TwitterURL,
			LinkedinURL:              employee.LinkedinURL,
			AccountTitle:             employee.AccountTitle,
			AccountNumber:            employee.AccountNumber,
			BankName:                 employee.BankName,
			Iban:                     employee.Iban,
			SwiftCode:                employee.SwiftCode,
			BankBranch:               employee.BankBranch,
			EmergencyContactFullName: employee.EmergencyContactFullName,
			EmergencyContactNumber:   employee.EmergencyContactNumber,
			EmergencyContactEmail:    employee.EmergencyContactEmail,
			EmergencyContactAddress:  employee.EmergencyContactAddress,
			CreatedAt:                employee.CreatedAt,
			UpdatedAt:                employee.UpdatedAt,
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
		result = db.Preload("Shift").
			Preload("Role").
			Preload("Department").
			Preload("Designation").
			First(&existingEmployee, "id = ?", employeeID)
		if result.Error != nil {
			return c.JSON(http.StatusNotFound, helper.ErrorResponse{Code: http.StatusNotFound, Message: "Employee not found"})
		}

		var updatedEmployee models.Employee
		if err := c.Bind(&updatedEmployee); err != nil {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"})
		}

		// Validate FirstName
		if updatedEmployee.FirstName != "" {
			if len(updatedEmployee.FirstName) < 3 || len(updatedEmployee.FirstName) > 30 || !regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(updatedEmployee.FirstName) {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "First name must be between 3 and 30 characters and contain only letters"})
			}
			existingEmployee.FirstName = updatedEmployee.FirstName
			existingEmployee.FullName = existingEmployee.FirstName + " " + existingEmployee.LastName // Update full name
		}

		// Validate LastName
		if updatedEmployee.LastName != "" {
			/*
				if len(updatedEmployee.LastName) < 3 || len(updatedEmployee.LastName) > 30 || !regexp.MustCompile(`^[a-zA-Z\s]+$`).MatchString(updatedEmployee.LastName) {
					return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Last name must be between 3 and 30 characters and contain only letters"})
				}
			*/
			existingEmployee.LastName = updatedEmployee.LastName
			existingEmployee.FullName = existingEmployee.FirstName + " " + existingEmployee.LastName // Update full name
		}

		// Validate ContactNumber
		if updatedEmployee.ContactNumber != "" {
			if len(updatedEmployee.ContactNumber) < 10 || len(updatedEmployee.ContactNumber) > 14 || !regexp.MustCompile(`^[0-9]+$`).MatchString(updatedEmployee.ContactNumber) {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Contact number must be between 10 and 14 digits and contain only numbers"})
			}
			existingEmployee.ContactNumber = updatedEmployee.ContactNumber
		}

		if updatedEmployee.Gender != "" {
			existingEmployee.Gender = updatedEmployee.Gender
		}

		if updatedEmployee.BirthdayDate != "" {
			startDate, err := time.Parse("2006-01-02", updatedEmployee.BirthdayDate)
			if err != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid StartDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingEmployee.BirthdayDate = startDate.Format("2006-01-02")
		}

		if updatedEmployee.IsActive != existingEmployee.IsActive {
			existingEmployee.IsActive = updatedEmployee.IsActive
		}

		// Validate Email
		if updatedEmployee.Email != "" {
			if !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(updatedEmployee.Email) {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid email format"})
			}
			existingEmployee.Email = updatedEmployee.Email
		}

		if updatedEmployee.Username != "" {
			// Validate username length
			if len(updatedEmployee.Username) < 5 || len(updatedEmployee.Username) > 15 || !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(updatedEmployee.Username) {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Username must be between 5 and 15 characters and contain only letters and numbers"})
			}
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

		if updatedEmployee.ShiftID != 0 {
			var officeShift models.Shift
			result = db.First(&officeShift, "id = ?", updatedEmployee.ShiftID)
			if result.Error != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid shift name. Shift not found."})
			}
			existingEmployee.ShiftID = updatedEmployee.ShiftID
		}
		if updatedEmployee.RoleID != 0 {
			var role models.Role
			result = db.First(&role, "id = ?", updatedEmployee.RoleID)
			if result.Error != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid role name. Role not found."})
			}
			existingEmployee.RoleID = updatedEmployee.RoleID
		}
		if updatedEmployee.DepartmentID != 0 {
			var department models.Department
			result = db.First(&department, "id = ?", updatedEmployee.DepartmentID)
			if result.Error != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid department name. Department not found."})
			}
			existingEmployee.DepartmentID = updatedEmployee.DepartmentID
		}
		if updatedEmployee.DesignationID != 0 {
			var designation models.Designation
			result = db.First(&designation, "id = ?", updatedEmployee.DesignationID)
			if result.Error != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid designation ID. Designation not found."})
			}
			existingEmployee.DesignationID = updatedEmployee.DesignationID
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

		if updatedEmployee.MaritalStatus != "" {
			existingEmployee.MaritalStatus = updatedEmployee.MaritalStatus
		}

		if updatedEmployee.Religion != "" {
			existingEmployee.Religion = updatedEmployee.Religion
		}

		if updatedEmployee.BloodGroup != "" {
			existingEmployee.BloodGroup = updatedEmployee.BloodGroup
		}

		if updatedEmployee.Nationality != "" {
			existingEmployee.Nationality = updatedEmployee.Nationality
		}

		if updatedEmployee.Citizenship != "" {
			existingEmployee.Citizenship = updatedEmployee.Citizenship
		}

		if updatedEmployee.BpjsKesehatan != "" {
			existingEmployee.BpjsKesehatan = updatedEmployee.BpjsKesehatan
		}

		if updatedEmployee.Address1 != "" {
			existingEmployee.Address1 = updatedEmployee.Address1
		}

		if updatedEmployee.Address2 != "" {
			existingEmployee.Address2 = updatedEmployee.Address2
		}

		if updatedEmployee.City != "" {
			existingEmployee.City = updatedEmployee.City
		}

		if updatedEmployee.StateProvince != "" {
			existingEmployee.StateProvince = updatedEmployee.StateProvince
		}

		if updatedEmployee.ZipPostalCode != "" {
			existingEmployee.ZipPostalCode = updatedEmployee.ZipPostalCode
		}
		if updatedEmployee.Bio != "" {
			existingEmployee.Bio = updatedEmployee.Bio
		}

		if updatedEmployee.FacebookURL != "" {
			existingEmployee.FacebookURL = updatedEmployee.FacebookURL
		}

		if updatedEmployee.InstagramURL != "" {
			existingEmployee.InstagramURL = updatedEmployee.InstagramURL
		}

		if updatedEmployee.TwitterURL != "" {
			existingEmployee.TwitterURL = updatedEmployee.TwitterURL
		}

		if updatedEmployee.LinkedinURL != "" {
			existingEmployee.LinkedinURL = updatedEmployee.LinkedinURL
		}

		if updatedEmployee.AccountTitle != "" {
			existingEmployee.AccountTitle = updatedEmployee.AccountTitle
		}

		if updatedEmployee.AccountNumber != "" {
			existingEmployee.AccountNumber = updatedEmployee.AccountNumber
		}

		if updatedEmployee.BankName != "" {
			existingEmployee.BankName = updatedEmployee.BankName
		}

		if updatedEmployee.Iban != "" {
			existingEmployee.Iban = updatedEmployee.Iban
		}

		if updatedEmployee.SwiftCode != "" {
			existingEmployee.SwiftCode = updatedEmployee.SwiftCode
		}

		if updatedEmployee.BankBranch != "" {
			existingEmployee.BankBranch = updatedEmployee.BankBranch
		}

		if updatedEmployee.EmergencyContactFullName != "" {
			existingEmployee.EmergencyContactFullName = updatedEmployee.EmergencyContactFullName
		}

		if updatedEmployee.EmergencyContactNumber != "" {
			existingEmployee.EmergencyContactNumber = updatedEmployee.EmergencyContactNumber
		}

		if updatedEmployee.EmergencyContactEmail != "" {
			existingEmployee.EmergencyContactEmail = updatedEmployee.EmergencyContactEmail
		}

		if updatedEmployee.EmergencyContactAddress != "" {
			existingEmployee.EmergencyContactAddress = updatedEmployee.EmergencyContactAddress
		}

		if err := db.Save(&existingEmployee).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update employee data"})
		}

		// Exclude PayrollInfo from the response
		employeeWithoutPayrollInfo := helper.EmployeeResponse{
			ID:                       existingEmployee.ID,
			PayrollID:                existingEmployee.PayrollID,
			FirstName:                existingEmployee.FirstName,
			LastName:                 existingEmployee.LastName,
			FullName:                 existingEmployee.FullName,
			ContactNumber:            existingEmployee.ContactNumber,
			Gender:                   existingEmployee.Gender,
			Email:                    existingEmployee.Email,
			Username:                 existingEmployee.Username,
			Password:                 existingEmployee.Password,
			ShiftID:                  existingEmployee.ShiftID,
			Shift:                    existingEmployee.Shift.ShiftName,
			RoleID:                   existingEmployee.RoleID,
			Role:                     existingEmployee.Role.RoleName,
			DepartmentID:             existingEmployee.DepartmentID,
			Department:               existingEmployee.Department.DepartmentName,
			DesignationID:            existingEmployee.DesignationID,
			Designation:              existingEmployee.Designation.DesignationName,
			BasicSalary:              existingEmployee.BasicSalary,
			HourlyRate:               existingEmployee.HourlyRate,
			PaySlipType:              existingEmployee.PaySlipType,
			IsActive:                 existingEmployee.IsActive,
			PaidStatus:               existingEmployee.PaidStatus,
			MaritalStatus:            existingEmployee.MaritalStatus,
			Religion:                 existingEmployee.Religion,
			BloodGroup:               existingEmployee.BloodGroup,
			Nationality:              existingEmployee.Nationality,
			Citizenship:              existingEmployee.Citizenship,
			BpjsKesehatan:            existingEmployee.BpjsKesehatan,
			Address1:                 existingEmployee.Address1,
			Address2:                 existingEmployee.Address2,
			City:                     existingEmployee.City,
			StateProvince:            existingEmployee.StateProvince,
			ZipPostalCode:            existingEmployee.ZipPostalCode,
			Bio:                      existingEmployee.Bio,
			FacebookURL:              existingEmployee.FacebookURL,
			InstagramURL:             existingEmployee.InstagramURL,
			TwitterURL:               existingEmployee.TwitterURL,
			LinkedinURL:              existingEmployee.LinkedinURL,
			AccountTitle:             existingEmployee.AccountTitle,
			AccountNumber:            existingEmployee.AccountNumber,
			BankName:                 existingEmployee.BankName,
			Iban:                     existingEmployee.Iban,
			SwiftCode:                existingEmployee.SwiftCode,
			BankBranch:               existingEmployee.BankBranch,
			EmergencyContactFullName: existingEmployee.EmergencyContactFullName,
			EmergencyContactNumber:   existingEmployee.EmergencyContactNumber,
			EmergencyContactEmail:    existingEmployee.EmergencyContactEmail,
			EmergencyContactAddress:  existingEmployee.EmergencyContactAddress,
			CreatedAt:                existingEmployee.CreatedAt,
			UpdatedAt:                existingEmployee.UpdatedAt,
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
		exitData.FullNameEmployee = employee.FirstName + " " + employee.LastName

		// Validasi apakah exit dengan ID yang diberikan ada
		var exit models.Exit
		result = db.First(&exit, exitData.ExitID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Exit not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		exitData.ExitName = exit.ExitName

		// Update status IsActive berdasarkan disable account
		if !exitData.DisableAccount {
			employee.IsActive = true
		} else {
			employee.IsActive = false
		}

		employee.IsExit = true

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

		// Handle search parameter
		searching := c.QueryParam("searching")

		var exitEmployees []models.ExitEmployee
		query := db.Order("id DESC").Offset(offset).Limit(perPage)

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("full_name_employee ILIKE ? OR exit_name ILIKE ? OR exit_interview ILIKE ? OR exit_date ILIKE ?", searchPattern, searchPattern, searchPattern, searchPattern)
		}

		if err := query.Find(&exitEmployees).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch ExitEmployee records"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var totalCount int64
		countQuery := db.Model(&models.ExitEmployee{})
		if searching != "" {
			searchPattern := "%" + searching + "%"
			countQuery = countQuery.Where("full_name_employee ILIKE ? OR exit_name ILIKE ? OR exit_interview ILIKE ? OR exit_date ILIKE ?", searchPattern, searchPattern, searchPattern, searchPattern)
		}
		countQuery.Count(&totalCount)

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
