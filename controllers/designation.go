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

type DesignationResponse struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	DepartmentID    uint      `json:"department_id"`
	DepartmentName  string    `json:"department_name"`
	DesignationName string    `json:"designation_name"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

func CreateDesignationByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var designation models.Designation
		if err := c.Bind(&designation); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if designation.DesignationName == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Designation name is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate RoleName using regexp
		if len(designation.DesignationName) < 5 || len(designation.DesignationName) > 30 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Designation name must be between 5 and 30 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if designation.DepartmentID == 0 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Department ID is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingDepartment models.Department
		result = db.Where("id = ?", designation.DepartmentID).First(&existingDepartment)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		designation.DepartmentName = existingDepartment.DepartmentName

		currentTime := time.Now()
		designation.CreatedAt = currentTime

		db.Create(&designation)

		// Create a response with DesignationResponse struct
		response := DesignationResponse{
			ID:              designation.ID,
			DepartmentID:    designation.DepartmentID,
			DepartmentName:  designation.DepartmentName,
			DesignationName: designation.DesignationName,
			Description:     designation.Description,
			CreatedAt:       designation.CreatedAt,
		}

		successResponse := map[string]interface{}{
			"Code":        http.StatusCreated,
			"Error":       false,
			"Message":     "Designation created successfully",
			"Designation": &response,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

/*
func CreateDesignationByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var designation models.Designation
		if err := c.Bind(&designation); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if designation.DesignationName == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Designation name is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate RoleName using regexp
		if len(designation.DesignationName) < 5 || len(designation.DesignationName) > 30 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Designation name must be between 5 and 30 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if designation.DepartmentID == 0 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Department ID is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingDepartment models.Department
		result = db.Where("id = ?", designation.DepartmentID).First(&existingDepartment)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Department not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		designation.DepartmentName = existingDepartment.DepartmentName

		currentTime := time.Now()
		designation.CreatedAt = currentTime

		db.Create(&designation)

		// Preload the Department data
		db.Preload("Department.Employee").First(&designation, designation.ID)

		successResponse := helper.Response{
			Code:        http.StatusCreated,
			Error:       false,
			Message:     "Designation created successfully",
			Designation: &designation,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}
*/

func GetAllDesignationsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var designations []models.Designation
		query := db.Preload("Department").Order("id DESC").Offset(offset).Limit(perPage)

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("designation_name ILIKE ?", searchPattern)
		}

		if err := query.Find(&designations).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch Designation records"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Batch fetch department names
		var departmentIDs []uint
		departmentMap := make(map[uint]string)
		for _, d := range designations {
			departmentIDs = append(departmentIDs, d.DepartmentID)
			departmentMap[d.DepartmentID] = ""
		}

		var departments []models.Department
		db.Where("id IN (?)", departmentIDs).Find(&departments)

		for _, dep := range departments {
			departmentMap[dep.ID] = dep.DepartmentName
		}

		// Assign department names to designations and update database
		tx := db.Begin()
		for i := range designations {
			if departmentName, ok := departmentMap[designations[i].DepartmentID]; ok {
				designations[i].DepartmentName = departmentName
			}
			if err := tx.Save(&designations[i]).Error; err != nil {
				tx.Rollback()
				errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Error saving designation data"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}
		tx.Commit()

		var totalCount int64
		countQuery := db.Model(&models.Designation{})
		if searching != "" {
			searchPattern := "%" + searching + "%"
			countQuery = countQuery.Where("designation_name ILIKE ?", searchPattern)
		}
		countQuery.Count(&totalCount)

		// Map to slice of DesignationResponse
		var designationResponses []DesignationResponse
		for _, d := range designations {
			response := DesignationResponse{
				ID:              d.ID,
				DepartmentID:    d.DepartmentID,
				DepartmentName:  d.DepartmentName,
				DesignationName: d.DesignationName,
				Description:     d.Description,
				CreatedAt:       d.CreatedAt,
			}
			designationResponses = append(designationResponses, response)
		}

		successResponse := map[string]interface{}{
			"code":         http.StatusOK,
			"error":        false,
			"message":      "Designations retrieved successfully",
			"designations": designationResponses,
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
func GetAllDesignationsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var designations []models.Designation
		query := db.Preload("Department").Order("id DESC").Offset(offset).Limit(perPage)

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("designation_name ILIKE ?", searchPattern)
		}

		if err := query.Find(&designations).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch Designation records"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Batch fetch department names
		var departmentIDs []uint
		departmentMap := make(map[uint]string)
		for _, d := range designations {
			departmentIDs = append(departmentIDs, d.DepartmentID)
			departmentMap[d.DepartmentID] = ""
		}

		var departments []models.Department
		db.Where("id IN (?)", departmentIDs).Find(&departments)

		for _, dep := range departments {
			departmentMap[dep.ID] = dep.DepartmentName
		}

		// Assign department names to designations
		for i, des := range designations {
			if departmentName, ok := departmentMap[des.DepartmentID]; ok {
				designations[i].DepartmentName = departmentName
			}
		}

		var totalCount int64
		countQuery := db.Model(&models.Designation{})
		if searching != "" {
			searchPattern := "%" + searching + "%"
			countQuery = countQuery.Where("designation_name ILIKE ?", searchPattern)
		}
		countQuery.Count(&totalCount)

		// Map to slice of DesignationResponse
		var designationResponses []DesignationResponse
		for _, d := range designations {
			response := DesignationResponse{
				ID:              d.ID,
				DepartmentID:    d.DepartmentID,
				DepartmentName:  d.DepartmentName,
				DesignationName: d.DesignationName,
				Description:     d.Description,
				CreatedAt:       d.CreatedAt,
			}
			designationResponses = append(designationResponses, response)
		}

		successResponse := map[string]interface{}{
			"code":         http.StatusOK,
			"error":        false,
			"message":      "Designations retrieved successfully",
			"designations": designationResponses,
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

/*
func GetAllDesignationsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var designations []models.Designation
		query := db.Order("id DESC").Offset(offset).Limit(perPage)

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("department_name ILIKE ? OR designation_name ILIKE ?", searchPattern, searchPattern)
		}

		if err := query.Find(&designations).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch Designation records"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var totalCount int64
		countQuery := db.Model(&models.Designation{})
		if searching != "" {
			searchPattern := "%" + searching + "%"
			countQuery = countQuery.Where("department_name ILIKE ? OR designation_name ILIKE ?", searchPattern, searchPattern)
		}
		countQuery.Count(&totalCount)

		// Map to slice of DesignationResponse
		var designationResponses []DesignationResponse
		for _, d := range designations {
			response := DesignationResponse{
				ID:              d.ID,
				DepartmentID:    d.DepartmentID,
				DepartmentName:  d.DepartmentName,
				DesignationName: d.DesignationName,
				Description:     d.Description,
				CreatedAt:       d.CreatedAt,
			}
			designationResponses = append(designationResponses, response)
		}

		successResponse := map[string]interface{}{
			"code":         http.StatusOK,
			"error":        false,
			"message":      "Designations retrieved successfully",
			"designations": designationResponses,
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

/*
func GetAllDesignationsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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


		var designations []models.Designation
		query := db.Preload("Department.Employee").Order("id DESC").Offset(offset).Limit(perPage)

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("department_name ILIKE ? OR designation_name ILIKE ?", searchPattern, searchPattern)
		}

		if err := query.Find(&designations).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch Designation records"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var totalCount int64
		countQuery := db.Model(&models.Designation{})
		if searching != "" {
			searchPattern := "%" + searching + "%"
			countQuery = countQuery.Where("department_name ILIKE ? OR designation_name ILIKE ?", searchPattern, searchPattern)
		}
		countQuery.Count(&totalCount)

		successResponse := map[string]interface{}{
			"code":         http.StatusOK,
			"error":        false,
			"message":      "Designations retrieved successfully",
			"designations": designations,
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

func GetDesignationByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		designationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid designation ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var designation models.Designation
		result = db.Preload("Department.Employee").Where("id = ?", designationID).First(&designation)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Designation not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Map to DesignationResponse
		designationResponse := DesignationResponse{
			ID:              designation.ID,
			DepartmentID:    designation.DepartmentID,
			DepartmentName:  designation.DepartmentName,
			DesignationName: designation.DesignationName,
			Description:     designation.Description,
			CreatedAt:       designation.CreatedAt,
		}

		successResponse := map[string]interface{}{
			"code":        http.StatusOK,
			"error":       false,
			"message":     "Designation retrieved successfully",
			"designation": designationResponse,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func GetDesignationByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		designationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid designation ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var designation models.Designation
		result = db.Where("id = ?", designationID).First(&designation)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Designation not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := helper.Response{
			Code:        http.StatusOK,
			Error:       false,
			Message:     "Designation retrieved successfully",
			Designation: &designation,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func UpdateDesignationByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		designationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid designation ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingDesignation models.Designation
		result = db.Where("id = ?", designationID).First(&existingDesignation)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Designation not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedDesignation models.Designation
		if err := c.Bind(&updatedDesignation); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedDesignation.DepartmentID != 0 {
			var department models.Department
			result = db.First(&department, "id = ?", updatedDesignation.DepartmentID)
			if result.Error != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Department not found"})
			}
			existingDesignation.DepartmentID = updatedDesignation.DepartmentID
			existingDesignation.DepartmentName = department.DepartmentName
		}

		if updatedDesignation.DesignationName != "" {
			if len(updatedDesignation.DesignationName) < 5 || len(updatedDesignation.DesignationName) > 30 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Designation name must be between 5 and 30 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingDesignation.DesignationName = updatedDesignation.DesignationName
		}

		if updatedDesignation.Description != "" {
			if len(updatedDesignation.Description) < 5 || len(updatedDesignation.Description) > 3000 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Designation description must be between 5 and 3000 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingDesignation.Description = updatedDesignation.Description
		}

		db.Save(&existingDesignation)

		// Map to DesignationResponse
		designationResponse := DesignationResponse{
			ID:              existingDesignation.ID,
			DepartmentID:    existingDesignation.DepartmentID,
			DepartmentName:  existingDesignation.DepartmentName,
			DesignationName: existingDesignation.DesignationName,
			Description:     existingDesignation.Description,
			CreatedAt:       existingDesignation.CreatedAt,
		}

		successResponse := map[string]interface{}{
			"Code":        http.StatusOK,
			"Error":       false,
			"Message":     "Designation updated successfully",
			"Designation": &designationResponse,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func UpdateDesignationByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		designationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid designation ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingDesignation models.Designation
		result = db.Where("id = ?", designationID).First(&existingDesignation)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Designation not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedDesignation models.Designation
		if err := c.Bind(&updatedDesignation); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedDesignation.DepartmentID != 0 {
			var department models.Department
			result = db.First(&department, "id = ?", updatedDesignation.DepartmentID)
			if result.Error != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Department not found"})
			}
			existingDesignation.DepartmentID = updatedDesignation.DepartmentID
			existingDesignation.DepartmentName = department.DepartmentName
		}

		if updatedDesignation.DesignationName != "" {
			if len(updatedDesignation.DesignationName) < 5 || len(updatedDesignation.DesignationName) > 30 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Designation name must be between 5 and 30 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingDesignation.DesignationName = updatedDesignation.DesignationName
		}

		if updatedDesignation.Description != "" {
			if len(updatedDesignation.Description) < 5 || len(updatedDesignation.Description) > 3000 {
				errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Designation description must be between 5 and 3000 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingDesignation.Description = updatedDesignation.Description
		}

		db.Save(&existingDesignation)

		successResponse := helper.Response{
			Code:        http.StatusOK,
			Error:       false,
			Message:     "Designation updated successfully",
			Designation: &existingDesignation,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func DeleteDesignationByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		designationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid designation ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingDesignation models.Designation
		result = db.Where("id = ?", designationID).First(&existingDesignation)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Designation not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// cek asosiasi dengan table employee
		var count int64
		db.Model(&models.Employee{}).Where("designation_id = ?", designationID).Count(&count)
		if count > 0 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Cannot delete designation because it is associated with one or more employees"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		db.Delete(&existingDesignation)

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Designation deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
