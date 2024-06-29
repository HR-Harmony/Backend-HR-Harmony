// controllers/employeeProfile.go

package controllers

import (
	"fmt"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func EmployeeProfile(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
		result := db.Preload("Shift").Preload("Role").Preload("Department").Preload("Designation").Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		employeeProfile := map[string]interface{}{
			"id":                          employee.ID,
			"first_name":                  employee.FirstName,
			"last_name":                   employee.LastName,
			"full_name":                   employee.FirstName + " " + employee.LastName,
			"contact_number":              employee.ContactNumber,
			"gender":                      employee.Gender,
			"email":                       employee.Email,
			"username":                    employee.Username,
			"shift_id":                    employee.ShiftID,
			"shift":                       employee.Shift.ShiftName,
			"role_id":                     employee.RoleID,
			"role":                        employee.Role.RoleName,
			"department_id":               employee.DepartmentID,
			"department":                  employee.Department.DepartmentName,
			"basic_salary":                employee.BasicSalary,
			"hourly_rate":                 employee.HourlyRate,
			"pay_slip_type":               employee.PaySlipType,
			"is_active":                   employee.IsActive,
			"paid_status":                 employee.PaidStatus,
			"marital_status":              employee.MaritalStatus,
			"religion":                    employee.Religion,
			"blood_group":                 employee.BloodGroup,
			"nationality":                 employee.Nationality,
			"citizenship":                 employee.Citizenship,
			"bpjs_kesehatan":              employee.BpjsKesehatan,
			"address1":                    employee.Address1,
			"address2":                    employee.Address2,
			"city":                        employee.City,
			"state_province":              employee.StateProvince,
			"zip_postal_code":             employee.ZipPostalCode,
			"bio":                         employee.Bio,
			"facebook_url":                employee.FacebookURL,
			"instagram_url":               employee.InstagramURL,
			"twitter_url":                 employee.TwitterURL,
			"linkedin_url":                employee.LinkedinURL,
			"account_title":               employee.AccountTitle,
			"account_number":              employee.AccountNumber,
			"bank_name":                   employee.BankName,
			"iban":                        employee.Iban,
			"swift_code":                  employee.SwiftCode,
			"bank_branch":                 employee.BankBranch,
			"emergency_contact_full_name": employee.EmergencyContactFullName,
			"emergency_contact_number":    employee.EmergencyContactNumber,
			"emergency_contact_email":     employee.EmergencyContactEmail,
			"emergency_contact_address":   employee.EmergencyContactAddress,
			"birthday_date":               employee.BirthdayDate,
			"created_at":                  employee.CreatedAt,
			"updated_at":                  employee.UpdatedAt,
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Employee profile retrieved successfully",
			"profile": employeeProfile,
		})
	}
}

func UpdateEmployeeProfile(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var existingEmployee models.Employee
		result := db.Where("username = ?", username).First(&existingEmployee)
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

		if updatedEmployee.BirthdayDate != "" {
			startDate, err := time.Parse("2006-01-02", updatedEmployee.BirthdayDate)
			if err != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid StartDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingEmployee.BirthdayDate = startDate.Format("2006-01-02")
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
		employeeWithoutPayrollInfo := map[string]interface{}{
			"ID":                       existingEmployee.ID,
			"PayrollID":                existingEmployee.PayrollID,
			"FirstName":                existingEmployee.FirstName,
			"LastName":                 existingEmployee.LastName,
			"ContactNumber":            existingEmployee.ContactNumber,
			"Gender":                   existingEmployee.Gender,
			"Email":                    existingEmployee.Email,
			"Username":                 existingEmployee.Username,
			"Password":                 existingEmployee.Password,
			"ShiftID":                  existingEmployee.ShiftID,
			"Shift":                    existingEmployee.Shift.ShiftName,
			"RoleID":                   existingEmployee.RoleID,
			"Role":                     existingEmployee.Role.RoleName,
			"DepartmentID":             existingEmployee.DepartmentID,
			"Department":               existingEmployee.Department.DepartmentName,
			"DesignationID":            existingEmployee.DesignationID,
			"Designation":              existingEmployee.Designation.DesignationName,
			"BasicSalary":              existingEmployee.BasicSalary,
			"HourlyRate":               existingEmployee.HourlyRate,
			"PaySlipType":              existingEmployee.PaySlipType,
			"IsActive":                 existingEmployee.IsActive,
			"PaidStatus":               existingEmployee.PaidStatus,
			"MaritalStatus":            existingEmployee.MaritalStatus,
			"Religion":                 existingEmployee.Religion,
			"BloodGroup":               existingEmployee.BloodGroup,
			"Nationality":              existingEmployee.Nationality,
			"Citizenship":              existingEmployee.Citizenship,
			"BpjsKesehatan":            existingEmployee.BpjsKesehatan,
			"Address1":                 existingEmployee.Address1,
			"Address2":                 existingEmployee.Address2,
			"City":                     existingEmployee.City,
			"StateProvince":            existingEmployee.StateProvince,
			"ZipPostalCode":            existingEmployee.ZipPostalCode,
			"Bio":                      existingEmployee.Bio,
			"FacebookURL":              existingEmployee.FacebookURL,
			"InstagramURL":             existingEmployee.InstagramURL,
			"TwitterURL":               existingEmployee.TwitterURL,
			"LinkedinURL":              existingEmployee.LinkedinURL,
			"AccountTitle":             existingEmployee.AccountTitle,
			"AccountNumber":            existingEmployee.AccountNumber,
			"BankName":                 existingEmployee.BankName,
			"Iban":                     existingEmployee.Iban,
			"SwiftCode":                existingEmployee.SwiftCode,
			"BankBranch":               existingEmployee.BankBranch,
			"EmergencyContactFullName": existingEmployee.EmergencyContactFullName,
			"EmergencyContactNumber":   existingEmployee.EmergencyContactNumber,
			"EmergencyContactEmail":    existingEmployee.EmergencyContactEmail,
			"EmergencyContactAddress":  existingEmployee.EmergencyContactAddress,
			"CreatedAt":                existingEmployee.CreatedAt,
			"UpdatedAt":                existingEmployee.UpdatedAt,
			"FullName":                 existingEmployee.FullName,
			"BirthdayDate":             existingEmployee.BirthdayDate,
		}
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Employee profile updated successfully",
			"data":    employeeWithoutPayrollInfo,
		}
		return c.JSON(http.StatusOK, successResponse)

	}
}

// UpdateEmployeePassword handles updating an employee's password by the employee themselves
func UpdateEmployeePassword(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve the employee from the database using the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			} else {
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch employee data"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		// Bind the new password and repeat password from the request body
		var newPassword struct {
			NewPassword    string `json:"new_password"`
			RepeatPassword string `json:"repeat_password"`
		}
		if err := c.Bind(&newPassword); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate the new password and repeat password
		if newPassword.NewPassword == "" || newPassword.RepeatPassword == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "New password and repeat password are required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if newPassword.NewPassword != newPassword.RepeatPassword {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "New password and repeat password do not match"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to hash password"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Update employee's password
		employee.Password = string(hashedPassword)
		if err := db.Save(&employee).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to update employee password"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Send password change notification to the employee
		go func(email, fullName, newPassword string) {
			if err := helper.SendPasswordChangeNotification(email, fullName, newPassword); err != nil {
				fmt.Println("Failed to send password change notification email:", err)
			}
		}(employee.Email, employee.FirstName+" "+employee.LastName, newPassword.NewPassword)

		// Respond with success
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Employee password updated successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
