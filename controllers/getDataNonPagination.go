package controllers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"strings"
)

// Admin

func GetAllShiftsByAdminNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var shifts []models.Shift

		// Handle search parameter
		searching := c.QueryParam("searching")
		if searching != "" {
			searchPattern := "%" + searching + "%"
			if err := db.Where("shift_name ILIKE ?", searchPattern).Find(&shifts).Error; err != nil {
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Error fetching shifts"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		} else {
			if err := db.Find(&shifts).Error; err != nil {
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Error fetching shifts"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		var totalCount int64
		db.Model(&models.Shift{}).Count(&totalCount)

		successResponse := map[string]interface{}{
			"code":        http.StatusOK,
			"error":       false,
			"message":     "Shifts retrieved successfully",
			"shifts":      shifts,
			"total_count": totalCount,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllRolesByAdminNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Handle search parameter
		searching := c.QueryParam("searching")

		var roles []models.Role
		query := db
		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("role_name ILIKE ?", searchPattern)
		}

		if err := query.Find(&roles).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Error fetching roles"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Roles retrieved successfully",
			"roles":   roles,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllDepartmentsByAdminNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Handle search parameters
		searching := c.QueryParam("searching")

		var departments []models.Department
		query := db

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("department_name ILIKE ? OR full_name ILIKE ?", searchPattern, searchPattern)
		}

		if err := query.Find(&departments).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch Department records"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":        http.StatusOK,
			"error":       false,
			"message":     "Departments retrieved successfully",
			"departments": departments,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllDesignationsByAdminNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Handle search parameters
		searching := c.QueryParam("searching")

		var designations []models.Designation
		query := db

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("department_name ILIKE ? OR designation_name ILIKE ?", searchPattern, searchPattern)
		}

		if err := query.Find(&designations).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch Designation records"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":         http.StatusOK,
			"error":        false,
			"message":      "Designations retrieved successfully",
			"designations": designations,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllEmployeesByAdminNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employees []models.Employee
		query := db.Where("is_client = ? AND is_exit = ?", false, false)

		searching := c.QueryParam("searching")
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
				ContactNumber:            emp.ContactNumber,
				Gender:                   emp.Gender,
				Email:                    emp.Email,
				BirthdayDate:             emp.BirthdayDate,
				Username:                 emp.Username,
				ShiftID:                  emp.ShiftID,
				Shift:                    emp.Shift,
				RoleID:                   emp.RoleID,
				Role:                     emp.Role,
				DepartmentID:             emp.DepartmentID,
				Department:               emp.Department,
				DesignationID:            emp.DesignationID,
				Designation:              emp.Designation,
				BasicSalary:              emp.BasicSalary,
				HourlyRate:               emp.HourlyRate,
				PaySlipType:              emp.PaySlipType,
				IsActive:                 *emp.IsActive,
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

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":      http.StatusOK,
			"error":     false,
			"message":   "All employees retrieved successfully",
			"employees": employeesResponse,
		})
	}
}

func GetAllExitStatusByAdminNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var exitStatuses []models.Exit
		db.Find(&exitStatuses)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Exit statuses retrieved successfully",
			"exits":   exitStatuses,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllProjectsByAdminNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var projects []models.Project
		db.Find(&projects)

		successResponse := map[string]interface{}{
			"Code":     http.StatusOK,
			"Error":    false,
			"Message":  "Projects retrieved successfully",
			"Projects": projects,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllClientsByAdminNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		searching := c.QueryParam("searching")

		query := db.Model(&models.Employee{}).Where("is_client = ?", true)
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where(
				"LOWER(full_name) LIKE ? OR LOWER(username) LIKE ? OR LOWER(contact_number) LIKE ? OR LOWER(gender) LIKE ? OR LOWER(country) LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
			)
		}

		var clientEmployees []struct {
			ID            uint   `json:"id"`
			FirstName     string `json:"first_name"`
			LastName      string `json:"last_name"`
			FullName      string `json:"full_name"`
			ContactNumber string `json:"contact_number"`
			Gender        string `json:"gender"`
			Email         string `json:"email"`
			Username      string `json:"username"`
			Country       string `json:"country"`
			IsActive      bool   `json:"is_active"`
		}
		if err := query.Select("id", "first_name", "last_name", "full_name", "contact_number", "gender", "email", "username", "country", "is_active").Find(&clientEmployees).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching client data"})
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Client data retrieved successfully",
			"data":    clientEmployees,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllGoalTypesByAdminNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var goalTypes []models.GoalType
		query := db.Model(&models.GoalType{})

		searching := c.QueryParam("searching")
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where("LOWER(goal_type) LIKE ?", searchPattern)
		}

		if err := query.Find(&goalTypes).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching goal types"})
		}

		successResponse := map[string]interface{}{
			"code":      http.StatusOK,
			"error":     false,
			"message":   "All goal types retrieved successfully",
			"goalTypes": goalTypes,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllTrainersByAdminNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		searching := c.QueryParam("searching")

		var trainers []models.Trainer
		query := db.Model(&trainers)
		if searching != "" {
			query = query.Where("LOWER(full_name) LIKE ? OR contact_number LIKE ? OR LOWER(email) LIKE ? OR LOWER(expertise) LIKE ?",
				"%"+strings.ToLower(searching)+"%",
				"%"+searching+"%",
				"%"+strings.ToLower(searching)+"%",
				"%"+strings.ToLower(searching)+"%",
			)
		}
		query.Find(&trainers)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Trainers fetched successfully",
			"data":    trainers,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllTrainingSkillsByAdminNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		searching := strings.ToLower(c.QueryParam("searching"))

		var trainingSkills []models.TrainingSkill
		query := db.Model(&trainingSkills)
		if searching != "" {
			query = query.Where("LOWER(training_skill) LIKE ?", "%"+searching+"%")
		}
		query.Find(&trainingSkills)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "TrainingSkills fetched successfully",
			"data":    trainingSkills,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllTrainingsByAdminNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		searching := c.QueryParam("searching")

		var trainings []models.Training
		query := db.Model(&trainings)
		if searching != "" {
			searching = strings.ToLower(searching)
			query = query.Where("LOWER(full_name_trainer) LIKE ? OR LOWER(training_skill) LIKE ? OR LOWER(full_name_employee) LIKE ? OR CAST(training_cost AS VARCHAR) LIKE ?",
				"%"+searching+"%",
				"%"+searching+"%",
				"%"+searching+"%",
				searching,
			)
		}
		query.Find(&trainings)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Trainings fetched successfully",
			"data":    trainings,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllLeaveRequestTypesByAdminNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var leaveRequestTypes []models.LeaveRequestType
		db.Find(&leaveRequestTypes)

		searching := c.QueryParam("searching")
		if searching != "" {
			var filteredLeaveRequestTypes []models.LeaveRequestType
			for _, lrt := range leaveRequestTypes {
				if strings.Contains(strings.ToLower(lrt.LeaveType), strings.ToLower(searching)) ||
					strings.Contains(strings.ToLower(fmt.Sprintf("%d", lrt.DaysPerYears)), strings.ToLower(searching)) ||
					strings.Contains(strings.ToLower(fmt.Sprintf("%t", lrt.IsRequiresApproval)), strings.ToLower(searching)) {
					filteredLeaveRequestTypes = append(filteredLeaveRequestTypes, lrt)
				}
			}
			leaveRequestTypes = filteredLeaveRequestTypes
		}

		successResponse := map[string]interface{}{
			"code":                http.StatusOK,
			"error":               false,
			"message":             "Leave request types retrieved successfully",
			"leave_request_types": leaveRequestTypes,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// Employee
func GetAllProjectsByEmployeeNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var projects []models.Project
		db.Find(&projects)

		successResponse := map[string]interface{}{
			"Code":     http.StatusOK,
			"Error":    false,
			"Message":  "Projects retrieved successfully",
			"Projects": projects,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllDepartmentsByEmployeeNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Handle search parameters
		searching := c.QueryParam("searching")

		var departments []models.Department
		query := db

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("department_name ILIKE ? OR full_name ILIKE ?", searchPattern, searchPattern)
		}

		if err := query.Find(&departments).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch Department records"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":        http.StatusOK,
			"error":       false,
			"message":     "Departments retrieved successfully",
			"departments": departments,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllClientsByEmployeeNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		searching := c.QueryParam("searching")

		query := db.Model(&models.Employee{}).Where("is_client = ?", true)
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where(
				"LOWER(full_name) LIKE ? OR LOWER(username) LIKE ? OR LOWER(contact_number) LIKE ? OR LOWER(gender) LIKE ? OR LOWER(country) LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
			)
		}

		var clientEmployees []struct {
			ID            uint   `json:"id"`
			FirstName     string `json:"first_name"`
			LastName      string `json:"last_name"`
			FullName      string `json:"full_name"`
			ContactNumber string `json:"contact_number"`
			Gender        string `json:"gender"`
			Email         string `json:"email"`
			Username      string `json:"username"`
			Country       string `json:"country"`
			IsActive      bool   `json:"is_active"`
		}
		if err := query.Select("id", "first_name", "last_name", "full_name", "contact_number", "gender", "email", "username", "country", "is_active").Find(&clientEmployees).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching client data"})
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Client data retrieved successfully",
			"data":    clientEmployees,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllLeaveRequestTypesByEmployeeNonPagination(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employeeUser models.Employee
		result := db.Where("username = ?", username).First(&employeeUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var leaveRequestTypes []models.LeaveRequestType
		db.Find(&leaveRequestTypes)

		searching := c.QueryParam("searching")
		if searching != "" {
			var filteredLeaveRequestTypes []models.LeaveRequestType
			for _, lrt := range leaveRequestTypes {
				if strings.Contains(strings.ToLower(lrt.LeaveType), strings.ToLower(searching)) ||
					strings.Contains(strings.ToLower(fmt.Sprintf("%d", lrt.DaysPerYears)), strings.ToLower(searching)) ||
					strings.Contains(strings.ToLower(fmt.Sprintf("%t", lrt.IsRequiresApproval)), strings.ToLower(searching)) {
					filteredLeaveRequestTypes = append(filteredLeaveRequestTypes, lrt)
				}
			}
			leaveRequestTypes = filteredLeaveRequestTypes
		}

		successResponse := map[string]interface{}{
			"code":                http.StatusOK,
			"error":               false,
			"message":             "Leave request types retrieved successfully",
			"leave_request_types": leaveRequestTypes,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
