package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
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
