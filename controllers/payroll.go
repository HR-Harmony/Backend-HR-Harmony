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

func GetAllEmployeesPayrollInfo(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve all employees' payroll information
		var employees []models.Employee
		result = db.Find(&employees)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve employees"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Create a response with required information
		var payrollInfoList []map[string]interface{}
		for _, employee := range employees {
			payrollInfo := map[string]interface{}{
				"payroll_id":   employee.PayrollID,
				"username":     employee.Username,
				"employee_id":  employee.ID,
				"payslip_type": employee.PaySlipType,
				"basic_salary": employee.BasicSalary,
				"hourly_rate":  employee.HourlyRate,
				"paid_status":  employee.PaidStatus,
			}
			payrollInfoList = append(payrollInfoList, payrollInfo)
		}

		// Respond with success
		successResponse := helper.Response{
			Code:        http.StatusOK,
			Error:       false,
			Message:     "Employee payroll information retrieved successfully",
			PayrollInfo: payrollInfoList,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdatePaidStatusByPayrollID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Extract PayrollID from the request
		payrollID := c.Param("payroll_id")

		var employee models.Employee
		if err := db.Preload("PayrollInfo").Where("payroll_id = ?", payrollID).First(&employee).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Update PaidStatus to true
		employee.PaidStatus = true
		db.Save(&employee)

		// Record payroll information
		payrollInfo := models.PayrollInfo{
			EmployeeID:  employee.ID,
			BasicSalary: employee.BasicSalary,
			PayslipType: employee.PaySlipType,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		db.Create(&payrollInfo)

		// Respond with success excluding PayrollInfo
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Paid status updated successfully",
			"employee": map[string]interface{}{
				"id":             employee.ID,
				"payroll_id":     employee.PayrollID,
				"first_name":     employee.FirstName,
				"last_name":      employee.LastName,
				"contact_number": employee.ContactNumber,
				"gender":         employee.Gender,
				"email":          employee.Email,
				"username":       employee.Username,
				"shift_id":       employee.ShiftID,
				"shift":          employee.Shift,
				"role_id":        employee.RoleID,
				"role":           employee.Role,
				"department_id":  employee.DepartmentID,
				"department":     employee.Department,
				"designation_id": employee.DesignationID,
				"designation":    employee.Designation,
				"basic_salary":   employee.BasicSalary,
				"hourly_rate":    employee.HourlyRate,
				"pay_slip_type":  employee.PaySlipType,
				"is_active":      employee.IsActive,
				"paid_status":    employee.PaidStatus,
				"created_at":     employee.CreatedAt,
				"updated_at":     employee.UpdatedAt,
			},
		}
		return c.JSON(http.StatusOK, successResponse)

	}
}

// GetAllPayrollInfo retrieves all payroll information
func GetAllPayrollHistory(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"code": http.StatusUnauthorized, "error": true, "message": "Authorization token is missing"})
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"code": http.StatusUnauthorized, "error": true, "message": "Invalid token format"})
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"code": http.StatusUnauthorized, "error": true, "message": "Invalid token"})
		}

		// Check if the user is an admin
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"code": http.StatusNotFound, "error": true, "message": "Admin user not found"})
		}

		if !adminUser.IsAdminHR {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"code": http.StatusForbidden, "error": true, "message": "Access denied"})
		}

		// Fetch all payroll information
		var payrollInfoList []models.PayrollInfo
		if err := db.Find(&payrollInfoList).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching payroll information"})
		}

		// Respond with success
		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "message": "Payroll information retrieved successfully", "payroll_info_list": payrollInfoList})
	}
}

func CreateAdvanceSalaryByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Parse the token to get admin's username
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

		// Bind the AdvanceSalary data from the request body
		var advanceSalary models.AdvanceSalary
		if err := c.Bind(&advanceSalary); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate AdvanceSalary data
		if advanceSalary.EmployeeID == 0 || advanceSalary.MonthAndYear == "" || advanceSalary.Amount == 0 || advanceSalary.Reason == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid data. Employee ID, Month and Year, Amount, and Reason are required fields"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if the employee with the given ID exists
		var employee models.Employee
		result = db.First(&employee, advanceSalary.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Set the FullnameEmployee
		advanceSalary.FullnameEmployee = employee.FullName
		advanceSalary.Emi = advanceSalary.MonthlyInstallmentAmt

		// Validate date format
		_, err = time.Parse("2006-01", advanceSalary.MonthAndYear)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid date format. Required format: yyyy-mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Create the AdvanceSalary in the database
		db.Create(&advanceSalary)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Advance Salary created successfully",
			"data":    advanceSalary,
		}
		return c.JSON(http.StatusCreated, successResponse)

	}
}

func GetAllAdvanceSalariesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Parse the token to get admin's username
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

		// Get query parameter for searching
		searching := c.QueryParam("searching")

		// Fetch advance salary data from the database with optional search filters
		var advanceSalaries []models.AdvanceSalary
		query := db.Model(&advanceSalaries)
		if searching != "" {
			query = query.Where("LOWER(fullname_employee) LIKE ? OR amount = ?", "%"+strings.ToLower(searching)+"%", helper.ParseStringToInt(searching))
		}
		query.Find(&advanceSalaries)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Advance Salary history retrieved successfully",
			"data":    advanceSalaries,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAdvanceSalaryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Parse the token to get admin's username
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

		// Extract ID parameter from the request
		id := c.Param("id")

		// Parse ID to uint
		advanceSalaryID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve advance salary data from the database by ID
		var advanceSalary models.AdvanceSalary
		result = db.First(&advanceSalary, advanceSalaryID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Advance Salary not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Advance Salary retrieved successfully",
			"data":    advanceSalary,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateAdvanceSalaryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Parse the token to get admin's username
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

		// Extract ID parameter from the request
		id := c.Param("id")

		// Parse ID to uint
		advanceSalaryID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve advance salary data from the database by ID
		var advanceSalary models.AdvanceSalary
		result = db.First(&advanceSalary, advanceSalaryID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Advance Salary not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Bind the updated data from the request body
		var updatedData models.AdvanceSalary
		if err := c.Bind(&updatedData); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update only the fields that are provided in the request body
		if updatedData.Amount != 0 {
			advanceSalary.Amount = updatedData.Amount
		}
		if updatedData.OneTimeDeduct != "" {
			advanceSalary.OneTimeDeduct = updatedData.OneTimeDeduct
		}
		if updatedData.MonthlyInstallmentAmt != 0 {
			advanceSalary.MonthlyInstallmentAmt = updatedData.MonthlyInstallmentAmt
		}
		if updatedData.Reason != "" {
			advanceSalary.Reason = updatedData.Reason
		}
		if updatedData.Emi != 0 {
			advanceSalary.Emi = updatedData.Emi
		}
		if updatedData.Paid != 0 {
			advanceSalary.Paid = updatedData.Paid
		}
		if updatedData.EmployeeID != 0 {
			// Retrieve employee data by ID
			var employee models.Employee
			result := db.First(&employee, updatedData.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			advanceSalary.EmployeeID = updatedData.EmployeeID
			advanceSalary.FullnameEmployee = employee.FullName
		}

		// Update the AdvanceSalary in the database
		db.Save(&advanceSalary)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Advance Salary updated successfully",
			"data":    advanceSalary,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteAdvanceSalaryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Parse the token to get admin's username
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

		// Extract ID parameter from the request
		id := c.Param("id")

		// Parse ID to uint
		advanceSalaryID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve advance salary data from the database by ID
		var advanceSalary models.AdvanceSalary
		result = db.First(&advanceSalary, advanceSalaryID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Advance Salary not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Delete the AdvanceSalary from the database
		db.Delete(&advanceSalary)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Advance Salary deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
