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

type DisciplinaryResponse struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	EmployeeID       uint       `json:"employee_id"`
	UsernameEmployee string     `json:"username_employee"`
	FullNameEmployee string     `json:"full_name_employee"`
	CaseID           uint       `json:"case_id"`
	CaseName         string     `json:"case_name"`
	Subject          string     `json:"subject"`
	CaseDate         string     `json:"case_date"`
	Description      string     `json:"description"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func CreateDisciplinaryByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var disciplinary models.Disciplinary
		if err := c.Bind(&disciplinary); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(disciplinary.Subject) < 5 || len(disciplinary.Subject) > 100 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Disciplinary subject must be between 5 and 100 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(disciplinary.Description) < 5 || len(disciplinary.Description) > 3000 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Disciplinary description must be between 5 and 3000 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		caseDate, err := time.Parse("2006-01-02", disciplinary.CaseDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid CaseDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		disciplinary.CaseDate = caseDate.Format("2006-01-02")

		if disciplinary.EmployeeID == 0 || disciplinary.CaseID == 0 || disciplinary.Subject == "" || disciplinary.CaseDate == "" || disciplinary.Description == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "All fields are required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingEmployee models.Employee
		result = db.First(&existingEmployee, disciplinary.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var existingCase models.Case
		result = db.First(&existingCase, disciplinary.CaseID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Case not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		disciplinary.UsernameEmployee = existingEmployee.Username
		disciplinary.FullNameEmployee = existingEmployee.FirstName + " " + existingEmployee.LastName
		disciplinary.CaseName = existingCase.CaseName

		currentTime := time.Now()
		disciplinary.CreatedAt = &currentTime

		db.Create(&disciplinary)

		// Mengirim notifikasi email kepada karyawan
		err = helper.SendDisciplinaryNotification(existingEmployee.Email, disciplinary.FullNameEmployee, disciplinary.CaseName, disciplinary.Subject, disciplinary.CaseDate, disciplinary.Description)
		if err != nil {
			fmt.Println("Failed to send disciplinary notification email:", err)
		}

		// Prepare response using DisciplinaryResponse struct
		response := DisciplinaryResponse{
			ID:               disciplinary.ID,
			EmployeeID:       disciplinary.EmployeeID,
			UsernameEmployee: disciplinary.UsernameEmployee,
			FullNameEmployee: disciplinary.FullNameEmployee,
			CaseID:           disciplinary.CaseID,
			CaseName:         disciplinary.CaseName,
			Subject:          disciplinary.Subject,
			CaseDate:         disciplinary.CaseDate,
			Description:      disciplinary.Description,
			CreatedAt:        disciplinary.CreatedAt,
			UpdatedAt:        disciplinary.UpdatedAt,
		}

		successResponse := map[string]interface{}{
			"Code":         http.StatusCreated,
			"Error":        false,
			"Message":      "Disciplinary data created successfully",
			"Disciplinary": &response,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

/*
func CreateDisciplinaryByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var disciplinary models.Disciplinary
		if err := c.Bind(&disciplinary); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(disciplinary.Subject) < 5 || len(disciplinary.Subject) > 100 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Disciplinary subject must be between 5 and 100 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(disciplinary.Description) < 5 || len(disciplinary.Description) > 3000 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Disciplinary description must be between 5 and 3000 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		caseDate, err := time.Parse("2006-01-02", disciplinary.CaseDate)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid CaseDate format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		disciplinary.CaseDate = caseDate.Format("2006-01-02")

		if disciplinary.EmployeeID == 0 || disciplinary.CaseID == 0 || disciplinary.Subject == "" || disciplinary.CaseDate == "" || disciplinary.Description == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "All fields are required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingEmployee models.Employee
		result = db.First(&existingEmployee, disciplinary.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var existingCase models.Case
		result = db.First(&existingCase, disciplinary.CaseID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Case not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		disciplinary.UsernameEmployee = existingEmployee.Username
		disciplinary.FullNameEmployee = existingEmployee.FirstName + " " + existingEmployee.LastName
		disciplinary.CaseName = existingCase.CaseName

		currentTime := time.Now()
		disciplinary.CreatedAt = &currentTime

		db.Create(&disciplinary)

		// Mengirim notifikasi email kepada karyawan
		err = helper.SendDisciplinaryNotification(existingEmployee.Email, disciplinary.FullNameEmployee, disciplinary.CaseName, disciplinary.Subject, disciplinary.CaseDate, disciplinary.Description)
		if err != nil {
			fmt.Println("Failed to send disciplinary notification email:", err)
		}

		successResponse := helper.Response{
			Code:         http.StatusCreated,
			Error:        false,
			Message:      "Disciplinary data created successfully",
			Disciplinary: &disciplinary,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}
*/

func GetAllDisciplinaryByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var disciplinaries []models.Disciplinary
		var totalCount int64

		// Fetch total count of records
		db.Model(&models.Disciplinary{}).Count(&totalCount)

		// Query paginated disciplinary data
		db.Order("id DESC").Offset(offset).Limit(perPage).Find(&disciplinaries)

		// Prepare response using DisciplinaryResponse struct for each disciplinary record
		var disciplinaryResponses []DisciplinaryResponse
		for _, disciplinary := range disciplinaries {
			disciplinaryResponse := DisciplinaryResponse{
				ID:               disciplinary.ID,
				EmployeeID:       disciplinary.EmployeeID,
				UsernameEmployee: disciplinary.UsernameEmployee,
				FullNameEmployee: disciplinary.FullNameEmployee,
				CaseID:           disciplinary.CaseID,
				CaseName:         disciplinary.CaseName,
				Subject:          disciplinary.Subject,
				CaseDate:         disciplinary.CaseDate,
				Description:      disciplinary.Description,
				CreatedAt:        disciplinary.CreatedAt,
				UpdatedAt:        disciplinary.UpdatedAt,
			}
			disciplinaryResponses = append(disciplinaryResponses, disciplinaryResponse)
		}

		// Construct success response
		successResponse := map[string]interface{}{
			"code":           http.StatusOK,
			"error":          false,
			"message":        "All disciplinary data retrieved successfully",
			"disciplinaries": disciplinaryResponses,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}

		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func GetAllDisciplinaryByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var disciplinaries []models.Disciplinary
		var totalCount int64
		db.Model(&models.Disciplinary{}).Count(&totalCount)
		db.Order("id DESC").Offset(offset).Limit(perPage).Find(&disciplinaries)

		successResponse := map[string]interface{}{
			"code":           http.StatusOK,
			"error":          false,
			"message":        "All disciplinary data retrieved successfully",
			"disciplinaries": disciplinaries,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func GetDisciplinaryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		disciplinaryIDStr := c.Param("id")
		disciplinaryID, err := strconv.ParseUint(disciplinaryIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid disciplinary ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var disciplinary models.Disciplinary
		result = db.First(&disciplinary, uint(disciplinaryID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Disciplinary data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		disciplinaryResponse := DisciplinaryResponse{
			ID:               disciplinary.ID,
			EmployeeID:       disciplinary.EmployeeID,
			UsernameEmployee: disciplinary.UsernameEmployee,
			FullNameEmployee: disciplinary.FullNameEmployee,
			CaseID:           disciplinary.CaseID,
			CaseName:         disciplinary.CaseName,
			Subject:          disciplinary.Subject,
			CaseDate:         disciplinary.CaseDate,
			Description:      disciplinary.Description,
			CreatedAt:        disciplinary.CreatedAt,
			UpdatedAt:        disciplinary.UpdatedAt,
		}

		successResponse := map[string]interface{}{
			"Code":         http.StatusOK,
			"Error":        false,
			"Message":      "Disciplinary data retrieved successfully",
			"Disciplinary": disciplinaryResponse,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func GetDisciplinaryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		disciplinaryIDStr := c.Param("id")
		disciplinaryID, err := strconv.ParseUint(disciplinaryIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid disciplinary ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var disciplinary models.Disciplinary
		result = db.First(&disciplinary, uint(disciplinaryID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Disciplinary data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := helper.Response{
			Code:         http.StatusOK,
			Error:        false,
			Message:      "Disciplinary data retrieved successfully",
			Disciplinary: &disciplinary,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func UpdateDisciplinaryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		disciplinaryIDStr := c.Param("id")
		disciplinaryID, err := strconv.ParseUint(disciplinaryIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid disciplinary ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var disciplinary models.Disciplinary
		result = db.First(&disciplinary, uint(disciplinaryID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Disciplinary data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedDisciplinary models.Disciplinary
		if err := c.Bind(&updatedDisciplinary); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedDisciplinary.EmployeeID != 0 {
			var existingEmployee models.Employee
			result = db.First(&existingEmployee, updatedDisciplinary.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			disciplinary.EmployeeID = updatedDisciplinary.EmployeeID
			disciplinary.UsernameEmployee = existingEmployee.Username
			disciplinary.FullNameEmployee = existingEmployee.FirstName + " " + existingEmployee.LastName
		}

		if updatedDisciplinary.CaseID != 0 {
			var existingCase models.Case
			result = db.First(&existingCase, updatedDisciplinary.CaseID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Case not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			disciplinary.CaseID = updatedDisciplinary.CaseID
			disciplinary.CaseName = existingCase.CaseName
		}

		if updatedDisciplinary.Subject != "" {
			if len(updatedDisciplinary.Subject) < 5 || len(updatedDisciplinary.Subject) > 100 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Disciplinary subject must be between 5 and 100 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			disciplinary.Subject = updatedDisciplinary.Subject
		}

		if updatedDisciplinary.CaseDate != "" {
			caseDate, err := time.Parse("2006-01-02", updatedDisciplinary.CaseDate)
			if err != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid CaseDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			disciplinary.CaseDate = caseDate.Format("2006-01-02")
		}

		if updatedDisciplinary.Description != "" {
			if len(updatedDisciplinary.Description) < 5 || len(updatedDisciplinary.Description) > 3000 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Disciplinary description must be between 5 and 3000 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			disciplinary.Description = updatedDisciplinary.Description
		}

		currentTime := time.Now()
		disciplinary.UpdatedAt = currentTime

		db.Save(&disciplinary)

		updatedDisciplinaryResponse := DisciplinaryResponse{
			ID:               disciplinary.ID,
			EmployeeID:       disciplinary.EmployeeID,
			UsernameEmployee: disciplinary.UsernameEmployee,
			FullNameEmployee: disciplinary.FullNameEmployee,
			CaseID:           disciplinary.CaseID,
			CaseName:         disciplinary.CaseName,
			Subject:          disciplinary.Subject,
			CaseDate:         disciplinary.CaseDate,
			Description:      disciplinary.Description,
			CreatedAt:        disciplinary.CreatedAt,
			UpdatedAt:        disciplinary.UpdatedAt,
		}

		successResponse := map[string]interface{}{
			"Code":         http.StatusOK,
			"Error":        false,
			"Message":      "Disciplinary data updated successfully",
			"Disciplinary": updatedDisciplinaryResponse,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func UpdateDisciplinaryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		disciplinaryIDStr := c.Param("id")
		disciplinaryID, err := strconv.ParseUint(disciplinaryIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid disciplinary ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var disciplinary models.Disciplinary
		result = db.First(&disciplinary, uint(disciplinaryID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Disciplinary data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedDisciplinary models.Disciplinary
		if err := c.Bind(&updatedDisciplinary); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedDisciplinary.EmployeeID != 0 {
			var existingEmployee models.Employee
			result = db.First(&existingEmployee, updatedDisciplinary.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			disciplinary.EmployeeID = updatedDisciplinary.EmployeeID
			disciplinary.UsernameEmployee = existingEmployee.Username
			disciplinary.FullNameEmployee = existingEmployee.FirstName + " " + existingEmployee.LastName
		}

		if updatedDisciplinary.CaseID != 0 {
			var existingCase models.Case
			result = db.First(&existingCase, updatedDisciplinary.CaseID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Case not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			disciplinary.CaseID = updatedDisciplinary.CaseID
			disciplinary.CaseName = existingCase.CaseName
		}

		if updatedDisciplinary.Subject != "" {
			if len(updatedDisciplinary.Subject) < 5 || len(updatedDisciplinary.Subject) > 100 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Disciplinary subject must be between 5 and 100 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			disciplinary.Subject = updatedDisciplinary.Subject
		}

		if updatedDisciplinary.CaseDate != "" {
			caseDate, err := time.Parse("2006-01-02", updatedDisciplinary.CaseDate)
			if err != nil {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid CaseDate format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			disciplinary.CaseDate = caseDate.Format("2006-01-02")
		}

		if updatedDisciplinary.Description != "" {
			if len(updatedDisciplinary.Description) < 5 || len(updatedDisciplinary.Description) > 3000 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Disciplinary description must be between 5 and 3000 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			disciplinary.Description = updatedDisciplinary.Description
		}

		currentTime := time.Now()
		disciplinary.UpdatedAt = currentTime

		db.Save(&disciplinary)

		successResponse := helper.Response{
			Code:         http.StatusOK,
			Error:        false,
			Message:      "Disciplinary data updated successfully",
			Disciplinary: &disciplinary,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func DeleteDisciplinaryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		disciplinaryIDStr := c.Param("id")
		disciplinaryID, err := strconv.ParseUint(disciplinaryIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid disciplinary ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var disciplinary models.Disciplinary
		result = db.First(&disciplinary, uint(disciplinaryID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Disciplinary data not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&disciplinary)

		successResponse := map[string]interface{}{
			"Code":    http.StatusOK,
			"Error":   false,
			"Message": "Disciplinary data deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
