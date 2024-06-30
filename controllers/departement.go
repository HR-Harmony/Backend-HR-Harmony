// controllers/createDepartement.go

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

type DepartmentResponse struct {
	ID             uint       `json:"id"`
	DepartmentName string     `json:"department_name"`
	EmployeeID     uint       `json:"employee_id"`
	FullName       string     `json:"full_name"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func CreateDepartemntsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var department models.Department
		if err := c.Bind(&department); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if department.DepartmentName == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Department name is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate DepartmentName
		if len(department.DepartmentName) < 5 || len(department.DepartmentName) > 30 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Department name must be between 5 and 30 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingDepartment models.Department
		result = db.Where("department_name = ?", department.DepartmentName).First(&existingDepartment)
		if result.Error == nil {
			errorResponse := helper.Response{Code: http.StatusConflict, Error: true, Message: "Department with this name already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		}

		var employee models.Employee
		result = db.First(&employee, "id = ?", department.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		department.FullName = employee.FullName

		currentTime := time.Now()
		department.CreatedAt = &currentTime

		db.Create(&department)

		// Create the response struct
		departmentResponse := DepartmentResponse{
			ID:             department.ID,
			DepartmentName: department.DepartmentName,
			EmployeeID:     department.EmployeeID,
			FullName:       department.FullName,
			CreatedAt:      department.CreatedAt,
			UpdatedAt:      department.UpdatedAt,
		}

		successResponse := map[string]interface{}{
			"Code":       http.StatusCreated,
			"Error":      false,
			"Message":    "Department created successfully",
			"Department": &departmentResponse,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

/*
func CreateDepartemntsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var department models.Department
		if err := c.Bind(&department); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if department.DepartmentName == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Department name is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate DepartmentName
		if len(department.DepartmentName) < 5 || len(department.DepartmentName) > 30 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Department name must be between 5 and 30 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingDepartment models.Department
		result = db.Where("department_name = ?", department.DepartmentName).First(&existingDepartment)
		if result.Error == nil {
			errorResponse := helper.Response{Code: http.StatusConflict, Error: true, Message: "Department with this name already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		}

		var employee models.Employee
		result = db.First(&employee, "id = ?", department.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		department.FullName = employee.FullName

		currentTime := time.Now()
		department.CreatedAt = &currentTime

		db.Create(&department)

		db.Preload("Employee").First(&department, department.ID)

		successResponse := helper.Response{
			Code:       http.StatusCreated,
			Error:      false,
			Message:    "Department created successfully",
			Department: &department,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}
*/

func GetAllDepartmentsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Handle search parameters
		searching := c.QueryParam("searching")

		var departments []models.Department
		query := db.Preload("Employee").Order("id DESC").Offset(offset).Limit(perPage)

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("department_name ILIKE ? OR full_name ILIKE ?", searchPattern, searchPattern)
		}

		if err := query.Find(&departments).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch Department records"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var totalCount int64
		countQuery := db.Model(&models.Department{})
		if searching != "" {
			searchPattern := "%" + searching + "%"
			countQuery = countQuery.Where("department_name ILIKE ? OR full_name ILIKE ?", searchPattern, searchPattern)
		}
		countQuery.Count(&totalCount)

		// Map departments to DepartmentResponse
		var departmentsResponse []DepartmentResponse
		for _, dep := range departments {
			departmentResp := DepartmentResponse{
				ID:             dep.ID,
				DepartmentName: dep.DepartmentName,
				EmployeeID:     dep.EmployeeID,
				FullName:       dep.FullName,
				CreatedAt:      dep.CreatedAt,
				UpdatedAt:      dep.UpdatedAt,
			}
			departmentsResponse = append(departmentsResponse, departmentResp)
		}

		successResponse := map[string]interface{}{
			"code":        http.StatusOK,
			"error":       false,
			"message":     "Departments retrieved successfully",
			"departments": departmentsResponse,
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
func GetAllDepartmentsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Handle search parameters
		searching := c.QueryParam("searching")


		var departments []models.Department
		query := db.Preload("Employee").Order("id DESC").Offset(offset).Limit(perPage)

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("department_name ILIKE ? OR full_name ILIKE ?", searchPattern, searchPattern)
		}

		if err := query.Find(&departments).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch Department records"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var totalCount int64
		countQuery := db.Model(&models.Department{})
		if searching != "" {
			searchPattern := "%" + searching + "%"
			countQuery = countQuery.Where("department_name ILIKE ? OR full_name ILIKE ?", searchPattern, searchPattern)
		}
		countQuery.Count(&totalCount)

		successResponse := map[string]interface{}{
			"code":        http.StatusOK,
			"error":       false,
			"message":     "Departments retrieved successfully",
			"departments": departments,
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

func GetDepartmentByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		departmentIDStr := c.Param("id")
		departmentID, err := strconv.ParseUint(departmentIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid department ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var department models.Department
		if err := db.First(&department, uint(departmentID)).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Construct the response with DepartmentResponse struct
		departmentResponse := DepartmentResponse{
			ID:             department.ID,
			DepartmentName: department.DepartmentName,
			EmployeeID:     department.EmployeeID,
			FullName:       department.FullName,
			CreatedAt:      department.CreatedAt,
			UpdatedAt:      department.UpdatedAt,
		}

		successResponse := map[string]interface{}{
			"Code":       http.StatusOK,
			"Error":      false,
			"Message":    "Department retrieved successfully",
			"Department": &departmentResponse,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func GetDepartmentByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		departmentIDStr := c.Param("id")
		departmentID, err := strconv.ParseUint(departmentIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid department ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Fetch the Department record by ID with preload on Employee
		var department models.Department
		if err := db.Preload("Employee").First(&department, uint(departmentID)).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := helper.Response{
			Code:       http.StatusOK,
			Error:      false,
			Message:    "Department retrieved successfully",
			Department: &department,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func EditDepartmentByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		departmentIDStr := c.Param("id")
		departmentID, err := strconv.ParseUint(departmentIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid department ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var department models.Department
		result = db.First(&department, uint(departmentID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updateData struct {
			DepartmentName string `json:"department_name"`
			EmployeeID     uint   `json:"employee_id"`
		}
		if err := c.Bind(&updateData); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updateData.DepartmentName != "" {
			if len(updateData.DepartmentName) < 5 || len(updateData.DepartmentName) > 30 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Department name must be between 5 and 30 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			department.DepartmentName = updateData.DepartmentName
		}

		if updateData.EmployeeID != 0 {
			var employee models.Employee
			result := db.First(&employee, "id = ?", updateData.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			department.EmployeeID = updateData.EmployeeID
			department.FullName = employee.FullName
		}

		department.UpdatedAt = time.Now()
		db.Save(&department)

		// Construct the response with DepartmentResponse struct
		departmentResponse := DepartmentResponse{
			ID:             department.ID,
			DepartmentName: department.DepartmentName,
			EmployeeID:     department.EmployeeID,
			FullName:       department.FullName,
			CreatedAt:      department.CreatedAt,
			UpdatedAt:      department.UpdatedAt,
		}

		successResponse := map[string]interface{}{
			"Code":       http.StatusOK,
			"Error":      false,
			"Message":    "Department updated successfully",
			"Department": &departmentResponse,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func EditDepartmentByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		departmentIDStr := c.Param("id")
		departmentID, err := strconv.ParseUint(departmentIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid department ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var department models.Department
		result = db.First(&department, uint(departmentID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updateData struct {
			DepartmentName string `json:"department_name"`
			EmployeeID     uint   `json:"employee_id"`
		}
		if err := c.Bind(&updateData); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updateData.DepartmentName != "" {
			if len(updateData.DepartmentName) < 5 || len(updateData.DepartmentName) > 30 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Department name must be between 5 and 30 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			department.DepartmentName = updateData.DepartmentName
		}

		if updateData.EmployeeID != 0 {
			var employee models.Employee
			result := db.First(&employee, "id = ?", updateData.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			department.EmployeeID = updateData.EmployeeID
			department.FullName = employee.FullName
		}

		department.UpdatedAt = time.Now()
		db.Save(&department)

		db.Preload("Employee").First(&department, department.ID)

		successResponse := helper.Response{
			Code:       http.StatusOK,
			Error:      false,
			Message:    "Department updated successfully",
			Department: &department,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func DeleteDepartmentByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		departmentIDStr := c.Param("id")
		departmentID, err := strconv.ParseUint(departmentIDStr, 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid department ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var department models.Department
		result = db.First(&department, uint(departmentID))
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&department)

		successResponse := map[string]interface{}{
			"Code":    http.StatusOK,
			"Error":   false,
			"Message": "Department deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
