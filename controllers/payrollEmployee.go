package controllers

import (
	"errors"
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

func GetPayrollInfoByEmployeeID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Searching parameter
		searching := c.QueryParam("searching")

		// Build the query
		query := db.Model(&models.PayrollInfo{}).Where("employee_id = ?", employee.ID)
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where(
				"LOWER(full_name_employee) LIKE ? OR LOWER(pay_slip_type) LIKE ?",
				searchPattern, searchPattern,
			)
		}

		// Retrieve payroll info data for the employee with pagination
		var payrollInfos []models.PayrollInfo
		result = query.Order("id DESC").Offset(offset).Limit(perPage).Find(&payrollInfos)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch payroll info data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Get total count of payroll info records for the employee
		var totalCount int64
		query.Count(&totalCount)

		// Return the payroll info data with pagination info
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Payroll info data retrieved successfully",
			"data":    payrollInfos,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetPayrollInfoByIDAndEmployeeID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		payrollInfoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid payrollInfo ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var payrollInfo models.PayrollInfo
		result = db.Where("id = ?", payrollInfoID).First(&payrollInfo)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "PayrollInfo ID not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch payrollInfo data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		if payrollInfo.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "PayrollInfo does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Payroll info data retrieved successfully",
			"data":    payrollInfo,
		})
	}
}

func CreateAdvanceSalaryByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var advanceSalary models.AdvanceSalary
		if err := c.Bind(&advanceSalary); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		advanceSalary.EmployeeID = employee.ID

		advanceSalary.FullnameEmployee = employee.FullName

		advanceSalary.Emi = advanceSalary.MonthlyInstallmentAmt

		advanceSalary.Status = "Pending"

		if advanceSalary.OneTimeDeduct == "Yes" {
			advanceSalary.MonthlyInstallmentAmt = advanceSalary.Amount
			advanceSalary.Emi = advanceSalary.MonthlyInstallmentAmt
		}

		_, err = time.Parse("2006-01", advanceSalary.MonthAndYear)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid date format. Required format: yyyy-mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		db.Create(&advanceSalary)

		// Mengirim notifikasi email kepada karyawan terkait
		err = helper.SendAdvanceSalaryNotification(employee.Email, advanceSalary.FullnameEmployee, advanceSalary.MonthAndYear, advanceSalary.Amount, advanceSalary.OneTimeDeduct, advanceSalary.MonthlyInstallmentAmt, advanceSalary.Reason)
		if err != nil {
			fmt.Println("Gagal mengirim email notifikasi advance salary:", err)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Advance Salary created successfully",
			"data":    advanceSalary,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAdvanceSalariesForEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
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

		// Searching parameter
		searching := c.QueryParam("searching")

		// Build the query
		query := db.Model(&models.AdvanceSalary{}).Where("employee_id = ?", employee.ID)
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where(
				"LOWER(fullname_employee) LIKE ? OR amount = ? OR emi = ? OR LOWER(status) LIKE ?",
				searchPattern, searching, searching, searchPattern,
			)
		}

		// Retrieve advance salaries for the employee with pagination
		var advanceSalaries []models.AdvanceSalary
		result = query.Offset(offset).Limit(perPage).Find(&advanceSalaries).Order("id DESC")
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch advance salaries"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Get total count of advance salaries records for the employee
		var totalCount int64
		query.Count(&totalCount)

		// Return the advance salaries data with pagination info
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Advance salaries retrieved successfully",
			"data":    advanceSalaries,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAdvanceSalaryByIDForEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		advanceSalaryIDStr := c.Param("id")
		advanceSalaryID, err := strconv.ParseUint(advanceSalaryIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid advance salary ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var advanceSalary models.AdvanceSalary
		result = db.Where("id = ?", advanceSalaryID).First(&advanceSalary)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Advance salary not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch advance salary data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		if advanceSalary.EmployeeID != employee.ID {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Advance salary does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Advance salary retrieved successfully",
			"data":    advanceSalary,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateAdvanceSalaryByIDForEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		advanceSalaryIDStr := c.Param("id")
		advanceSalaryID, err := strconv.ParseUint(advanceSalaryIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid advance salary ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var advanceSalary models.AdvanceSalary
		result = db.Where("id = ?", advanceSalaryID).First(&advanceSalary)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Advance salary not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch advance salary data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		if advanceSalary.EmployeeID != employee.ID {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Advance salary does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
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

		db.Save(&advanceSalary)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Advance salary updated successfully",
			"data":    advanceSalary,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteAdvanceSalaryByIDForEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		advanceSalaryIDStr := c.Param("id")
		advanceSalaryID, err := strconv.ParseUint(advanceSalaryIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid advance salary ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var advanceSalary models.AdvanceSalary
		result = db.Where("id = ?", advanceSalaryID).First(&advanceSalary)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Advance salary not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch advance salary data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		if advanceSalary.EmployeeID != employee.ID {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Advance salary does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		db.Delete(&advanceSalary)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Advance salary deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func CreateRequestLoanByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var requestLoan models.RequestLoan
		if err := c.Bind(&requestLoan); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		requestLoan.EmployeeID = employee.ID
		requestLoan.FullnameEmployee = employee.FullName

		if requestLoan.MonthAndYear == "" || requestLoan.Amount == 0 || requestLoan.Reason == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid data. Month and Year, Amount, and Reason are required fields"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		requestLoan.Status = "Pending"
		requestLoan.Emi = requestLoan.MonthlyInstallmentAmt
		requestLoan.Remaining = requestLoan.Amount - requestLoan.Paid

		if requestLoan.OneTimeDeduct == "Yes" {
			requestLoan.MonthlyInstallmentAmt = requestLoan.Amount
			requestLoan.Emi = requestLoan.MonthlyInstallmentAmt
		}

		_, err = time.Parse("2006-01", requestLoan.MonthAndYear)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid date format. Required format: yyyy-mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		db.Create(&requestLoan)

		// Mengirim notifikasi email kepada karyawan
		err = helper.SendRequestLoanNotification(employee.Email, employee.FullName, requestLoan.MonthAndYear, requestLoan.Amount, requestLoan.OneTimeDeduct, requestLoan.MonthlyInstallmentAmt, requestLoan.Reason)
		if err != nil {
			fmt.Println("Failed to send request loan notification email:", err)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Request Loan created successfully",
			"data":    requestLoan,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllRequestLoanByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
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

		// Searching parameter
		searching := c.QueryParam("searching")

		// Build the query
		query := db.Model(&models.RequestLoan{}).Where("employee_id = ?", employee.ID)
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where(
				"LOWER(fullname_employee) LIKE ? OR amount = ? OR emi = ? OR LOWER(status) LIKE ?",
				searchPattern, searching, searching, searchPattern,
			)
		}

		// Retrieve request loans for the employee with pagination
		var requestLoans []models.RequestLoan
		result = query.Order("id DESC").Offset(offset).Limit(perPage).Find(&requestLoans)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch request loans"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Get total count of request loans records for the employee
		var totalCount int64
		query.Count(&totalCount)

		// Return the request loans data with pagination info
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Request loans retrieved successfully",
			"data":    requestLoans,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetRequestLoanByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		requestLoanIDStr := c.Param("id")
		requestLoanID, err := strconv.ParseUint(requestLoanIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request loan ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var requestLoan models.RequestLoan
		result = db.Where("id = ?", requestLoanID).First(&requestLoan)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Request loan not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			} else {
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve request loan"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		if requestLoan.EmployeeID != employee.ID {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Request loan does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Request loan retrieved successfully",
			"data":    requestLoan,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateRequestLoanByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		requestLoanIDStr := c.Param("id")
		requestLoanID, err := strconv.ParseUint(requestLoanIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request loan ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var requestLoan models.RequestLoan
		result = db.Where("id = ?", requestLoanID).First(&requestLoan)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Request loan not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			} else {
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve request loan"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		if requestLoan.EmployeeID != employee.ID {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Request loan does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
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
			requestLoan.Remaining = updatedData.Amount - requestLoan.Paid
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

		db.Save(&requestLoan)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Request loan updated successfully",
			"data":    requestLoan,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteRequestLoanByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		requestLoanIDStr := c.Param("id")
		requestLoanID, err := strconv.ParseUint(requestLoanIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request loan ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var requestLoan models.RequestLoan
		result = db.Where("id = ?", requestLoanID).First(&requestLoan)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Request loan not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			} else {
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve request loan"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		if requestLoan.EmployeeID != employee.ID {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Request loan does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		db.Delete(&requestLoan)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Request loan deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
