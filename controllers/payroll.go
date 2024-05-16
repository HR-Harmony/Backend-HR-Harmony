package controllers

import (
	"fmt"
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

		var employees []models.Employee
		result = db.Where("is_client = ?", false).Offset(offset).Limit(perPage).Find(&employees)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve employees"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var payrollInfoList []map[string]interface{}
		for _, employee := range employees {
			payrollInfo := map[string]interface{}{
				"payroll_id":   employee.PayrollID,
				"username":     employee.Username,
				"full_name":    employee.FullName,
				"employee_id":  employee.ID,
				"payslip_type": employee.PaySlipType,
				"basic_salary": employee.BasicSalary,
				"hourly_rate":  employee.HourlyRate,
				"paid_status":  employee.PaidStatus,
			}
			payrollInfoList = append(payrollInfoList, payrollInfo)
		}

		var totalCount int64
		db.Model(&models.Employee{}).Where("is_client = ?", false).Count(&totalCount)

		successResponse := map[string]interface{}{
			"Code":        http.StatusOK,
			"Error":       false,
			"Message":     "Employee payroll information retrieved successfully",
			"PayrollInfo": payrollInfoList,
			"Pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdatePaidStatusByPayrollID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		payrollID := c.Param("payroll_id")

		var employee models.Employee
		if err := db.Preload("PayrollInfo").Where("payroll_id = ?", payrollID).First(&employee).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		employee.PaidStatus = true
		db.Save(&employee)

		payrollInfo := models.PayrollInfo{
			EmployeeID:  employee.ID,
			BasicSalary: employee.BasicSalary,
			PayslipType: employee.PaySlipType,
			PaidStatus:  employee.PaidStatus,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		db.Create(&payrollInfo)

		go func(email, fullName string, basicSalary float64) {
			if err := helper.SendSalaryTransferNotification(email, fullName, basicSalary); err != nil {
				fmt.Println("Failed to send salary transfer notification email:", err)
			}
		}(employee.Email, employee.FirstName+" "+employee.LastName, employee.BasicSalary)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Paid status updated successfully",
			"employee": map[string]interface{}{
				"id":             employee.ID,
				"payroll_id":     employee.PayrollID,
				"first_name":     employee.FirstName,
				"last_name":      employee.LastName,
				"full_name":      employee.FullName,
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

func GetAllPayrollHistory(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"code": http.StatusNotFound, "error": true, "message": "Admin user not found"})
		}

		if !adminUser.IsAdminHR {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"code": http.StatusForbidden, "error": true, "message": "Access denied"})
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

		var payrollInfoList []models.PayrollInfo
		if err := db.Offset(offset).Limit(perPage).Find(&payrollInfoList).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching payroll information"})
		}

		var totalCount int64
		db.Model(&models.PayrollInfo{}).Count(&totalCount)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":              http.StatusOK,
			"error":             false,
			"message":           "Payroll information retrieved successfully",
			"payroll_info_list": payrollInfoList,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		})
	}
}

func CreateAdvanceSalaryByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var advanceSalary models.AdvanceSalary
		if err := c.Bind(&advanceSalary); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if advanceSalary.EmployeeID == 0 || advanceSalary.MonthAndYear == "" || advanceSalary.Amount == 0 || advanceSalary.Reason == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid data. Employee ID, Month and Year, Amount, and Reason are required fields"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var employee models.Employee
		result = db.First(&employee, advanceSalary.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		advanceSalary.FullnameEmployee = employee.FullName
		advanceSalary.Emi = advanceSalary.MonthlyInstallmentAmt

		advanceSalary.Status = "Pending"

		if advanceSalary.OneTimeDeduct == "Yes" {
			advanceSalary.MonthlyInstallmentAmt = advanceSalary.Amount
			advanceSalary.Emi = advanceSalary.MonthlyInstallmentAmt
		}

		advanceSalary.Emi = advanceSalary.MonthlyInstallmentAmt

		_, err = time.Parse("2006-01", advanceSalary.MonthAndYear)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid date format. Required format: yyyy-mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		db.Create(&advanceSalary)

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

		var advanceSalaries []models.AdvanceSalary
		query := db.Model(&advanceSalaries)
		if searching != "" {
			query = query.Where("LOWER(fullname_employee) LIKE ? OR amount = ?", "%"+strings.ToLower(searching)+"%", helper.ParseStringToInt(searching))
		}
		query.Offset(offset).Limit(perPage).Find(&advanceSalaries)

		var totalCount int64
		db.Model(&models.AdvanceSalary{}).Count(&totalCount)

		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Advance Salary history retrieved successfully",
			"data":       advanceSalaries,
			"pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAdvanceSalaryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		id := c.Param("id")

		advanceSalaryID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var advanceSalary models.AdvanceSalary
		result = db.First(&advanceSalary, advanceSalaryID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Advance Salary not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

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

		id := c.Param("id")

		advanceSalaryID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var advanceSalary models.AdvanceSalary
		result = db.First(&advanceSalary, advanceSalaryID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Advance Salary not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedData models.AdvanceSalary
		if err := c.Bind(&updatedData); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedData.Amount != 0 {
			advanceSalary.Amount = updatedData.Amount
			advanceSalary.Emi = updatedData.Amount
			advanceSalary.MonthlyInstallmentAmt = updatedData.Amount
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

		if updatedData.Status != "" {
			advanceSalary.Status = updatedData.Status
		}

		if updatedData.EmployeeID != 0 {
			var employee models.Employee
			result := db.First(&employee, updatedData.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			advanceSalary.EmployeeID = updatedData.EmployeeID
			advanceSalary.FullnameEmployee = employee.FullName
		}

		db.Save(&advanceSalary)

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

		id := c.Param("id")

		advanceSalaryID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var advanceSalary models.AdvanceSalary
		result = db.First(&advanceSalary, advanceSalaryID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Advance Salary not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&advanceSalary)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Advance Salary deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func CreateRequestLoanByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var requestLoan models.RequestLoan
		if err := c.Bind(&requestLoan); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if requestLoan.EmployeeID == 0 || requestLoan.MonthAndYear == "" || requestLoan.Amount == 0 || requestLoan.Reason == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid data. Employee ID, Month and Year, Amount, and Reason are required fields"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var employee models.Employee
		result = db.First(&employee, requestLoan.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		requestLoan.FullnameEmployee = employee.FullName
		requestLoan.Emi = requestLoan.MonthlyInstallmentAmt

		requestLoan.Status = "Pending"

		requestLoan.Remaining = requestLoan.Amount - requestLoan.Paid

		if requestLoan.OneTimeDeduct == "Yes" {
			requestLoan.MonthlyInstallmentAmt = requestLoan.Amount
		}

		_, err = time.Parse("2006-01", requestLoan.MonthAndYear)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid date format. Required format: yyyy-mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		db.Create(&requestLoan)

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Request Loan created successfully",
			"data":    requestLoan,
		}
		return c.JSON(http.StatusCreated, successResponse)

	}
}

func GetAllRequestLoanByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var requestLoan []models.RequestLoan
		query := db.Model(&requestLoan)
		if searching != "" {
			query = query.Where("LOWER(fullname_employee) LIKE ? OR amount = ?", "%"+strings.ToLower(searching)+"%", helper.ParseStringToInt(searching))
		}
		query.Offset(offset).Limit(perPage).Find(&requestLoan)

		var totalCount int64
		db.Model(&models.RequestLoan{}).Count(&totalCount)

		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Request Loan history retrieved successfully",
			"data":       requestLoan,
			"pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetRequestLoanByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		id := c.Param("id")

		requestLoanID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var requestLoan models.RequestLoan
		result = db.First(&requestLoan, requestLoanID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Request Loan not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Request Loan retrieved successfully",
			"data":    requestLoan,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateRequestLoanByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		id := c.Param("id")

		requestLoanID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var requestLoan models.RequestLoan
		result = db.First(&requestLoan, requestLoanID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Request Loan not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedData models.RequestLoan
		if err := c.Bind(&updatedData); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedData.Amount != 0 {
			requestLoan.Amount = updatedData.Amount
			requestLoan.Emi = updatedData.Amount
			requestLoan.MonthlyInstallmentAmt = updatedData.Amount
		}
		if updatedData.OneTimeDeduct != "" {
			requestLoan.OneTimeDeduct = updatedData.OneTimeDeduct
		}
		if updatedData.MonthlyInstallmentAmt != 0 {
			requestLoan.MonthlyInstallmentAmt = updatedData.MonthlyInstallmentAmt
		}
		if updatedData.Reason != "" {
			requestLoan.Reason = updatedData.Reason
		}
		if updatedData.Emi != 0 {
			requestLoan.Emi = updatedData.Emi
		}
		if updatedData.Paid != 0 {
			requestLoan.Paid = updatedData.Paid
			requestLoan.Remaining = requestLoan.Amount - updatedData.Paid
		}

		if updatedData.Status != "" {
			requestLoan.Status = updatedData.Status
		}

		if updatedData.EmployeeID != 0 {
			var employee models.Employee
			result := db.First(&employee, updatedData.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			requestLoan.EmployeeID = updatedData.EmployeeID
			requestLoan.FullnameEmployee = employee.FullName
		}

		db.Save(&requestLoan)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Request Loan updated successfully",
			"data":    requestLoan,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteRequestLoanByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		id := c.Param("id")

		requestLoanID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var requestLoan models.RequestLoan
		result = db.First(&requestLoan, requestLoanID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Request Loan not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&requestLoan)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Request Loan deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
