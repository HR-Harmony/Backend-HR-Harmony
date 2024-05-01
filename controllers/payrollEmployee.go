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
	"time"
)

// GetPayrollInfoByEmployeeID mengambil data riwayat slip gaji milik karyawan berdasarkan ID karyawan
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

		// Retrieve payroll info data for the employee
		var payrollInfos []models.PayrollInfo
		db.Where("employee_id = ?", employee.ID).Find(&payrollInfos)

		// Return the payroll info data
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Payroll info data retrieved successfully",
			"data":    payrollInfos,
		})
	}
}

// CreateAdvanceSalaryForEmployee memungkinkan karyawan untuk menambahkan data advance salary untuk dirinya sendiri
func CreateAdvanceSalaryForEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Parse the token to get employee's username
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Bind the AdvanceSalary data from the request body
		var advanceSalary models.AdvanceSalary
		if err := c.Bind(&advanceSalary); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Set EmployeeID to the ID of the authenticated employee
		advanceSalary.EmployeeID = employee.ID

		// Set the FullnameEmployee
		advanceSalary.FullnameEmployee = employee.FullName

		advanceSalary.Emi = advanceSalary.MonthlyInstallmentAmt

		advanceSalary.Status = "Pending"

		if advanceSalary.OneTimeDeduct == "Yes" {
			advanceSalary.MonthlyInstallmentAmt = advanceSalary.Amount
			advanceSalary.Emi = advanceSalary.MonthlyInstallmentAmt
		}

		// Validate date format
		_, err = time.Parse("2006-01", advanceSalary.MonthAndYear)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid date format. Required format: yyyy-mm"}
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

// GetAdvanceSalariesForEmployee memungkinkan karyawan untuk melihat semua data advance salary miliknya
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

		// Parse the token to get employee's username
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Retrieve all advance salaries for the authenticated employee
		var advanceSalaries []models.AdvanceSalary
		result = db.Where("employee_id = ?", employee.ID).Find(&advanceSalaries)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch advance salaries"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Advance salaries retrieved successfully",
			"data":    advanceSalaries,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetAdvanceSalaryByIDForEmployee memungkinkan karyawan untuk melihat data advance salary miliknya berdasarkan ID advance salary
func GetAdvanceSalaryByIDForEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Parse the token to get employee's username
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Retrieve advance salary ID from path parameter
		advanceSalaryIDStr := c.Param("id")
		advanceSalaryID, err := strconv.ParseUint(advanceSalaryIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid advance salary ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the advance salary
		var advanceSalary models.AdvanceSalary
		result = db.Where("id = ?", advanceSalaryID).First(&advanceSalary)
		if result.Error != nil {
			// Check if advance salary not found
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Advance salary not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch advance salary data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Check if advance salary does not belong to the employee
		if advanceSalary.EmployeeID != employee.ID {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Advance salary does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Advance salary retrieved successfully",
			"data":    advanceSalary,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// CreateRequestLoanByEmployee memungkinkan karyawan untuk menambahkan data requestLoan untuk dirinya sendiri
func CreateRequestLoanByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Parse the token to get employee's username
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Bind the RequestLoan data from the request body
		var requestLoan models.RequestLoan
		if err := c.Bind(&requestLoan); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Set EmployeeID to the ID of the authenticated employee
		requestLoan.EmployeeID = employee.ID
		requestLoan.FullnameEmployee = employee.FullName

		// Validate RequestLoan data
		if requestLoan.MonthAndYear == "" || requestLoan.Amount == 0 || requestLoan.Reason == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid data. Month and Year, Amount, and Reason are required fields"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Set default values for some fields
		requestLoan.Status = "Pending"
		requestLoan.Emi = requestLoan.MonthlyInstallmentAmt
		requestLoan.Remaining = requestLoan.Amount - requestLoan.Paid

		// Set Monthly Installment Amount based on One Time Deduct
		if requestLoan.OneTimeDeduct == "Yes" {
			requestLoan.MonthlyInstallmentAmt = requestLoan.Amount
		}

		// Validate date format
		_, err = time.Parse("2006-01", requestLoan.MonthAndYear)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid date format. Required format: yyyy-mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Create the RequestLoan in the database
		db.Create(&requestLoan)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Request Loan created successfully",
			"data":    requestLoan,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

// GetRequestLoansForEmployee memungkinkan karyawan untuk melihat data requestLoan miliknya sendiri
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

		// Parse the token to get employee's username
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Retrieve request loans for the employee
		var requestLoans []models.RequestLoan
		db.Where("employee_id = ?", employee.ID).Find(&requestLoans)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Request Loans retrieved successfully",
			"data":    requestLoans,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// GetRequestLoanByIDForEmployee memungkinkan karyawan untuk melihat data requestLoan miliknya berdasarkan ID requestLoan
func GetRequestLoanByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Parse the token to get employee's username
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Retrieve request loan ID from path parameter
		requestLoanIDStr := c.Param("id")
		requestLoanID, err := strconv.ParseUint(requestLoanIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request loan ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the request loan
		var requestLoan models.RequestLoan
		result = db.Where("id = ?", requestLoanID).First(&requestLoan)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// Request loan not found
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Request loan not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			} else {
				// Other database error
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve request loan"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		// Check if the request loan belongs to the employee
		if requestLoan.EmployeeID != employee.ID {
			// Request loan does not belong to the employee
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Request loan does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Request loan retrieved successfully",
			"data":    requestLoan,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// UpdateRequestLoanByIDForEmployee memungkinkan karyawan untuk mengubah data request loan miliknya berdasarkan ID request loan
func UpdateRequestLoanByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Parse the token to get employee's username
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Retrieve request loan ID from path parameter
		requestLoanIDStr := c.Param("id")
		requestLoanID, err := strconv.ParseUint(requestLoanIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request loan ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the request loan
		var requestLoan models.RequestLoan
		result = db.Where("id = ?", requestLoanID).First(&requestLoan)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// Request loan not found
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Request loan not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			} else {
				// Other database error
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve request loan"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		// Check if the request loan belongs to the employee
		if requestLoan.EmployeeID != employee.ID {
			// Request loan does not belong to the employee
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Request loan does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Bind the updated data from the request body
		var updatedData models.RequestLoan
		if err := c.Bind(&updatedData); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update only the fields that are provided in the request body
		if updatedData.Amount != 0 {
			requestLoan.Amount = updatedData.Amount
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

		// Update the RequestLoan in the database
		db.Save(&requestLoan)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Request loan updated successfully",
			"data":    requestLoan,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// DeleteRequestLoanByIDForEmployee memungkinkan karyawan untuk menghapus data request loan miliknya berdasarkan ID request loan
func DeleteRequestLoanByIDByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Parse the token to get employee's username
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Retrieve request loan ID from path parameter
		requestLoanIDStr := c.Param("id")
		requestLoanID, err := strconv.ParseUint(requestLoanIDStr, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request loan ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve the request loan
		var requestLoan models.RequestLoan
		result = db.Where("id = ?", requestLoanID).First(&requestLoan)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				// Request loan not found
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Request loan not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			} else {
				// Other database error
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve request loan"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		// Check if the request loan belongs to the employee
		if requestLoan.EmployeeID != employee.ID {
			// Request loan does not belong to the employee
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Request loan does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Delete the request loan from the database
		db.Delete(&requestLoan)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Request loan deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
