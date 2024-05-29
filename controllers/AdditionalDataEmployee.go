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
)

func GetAllDepartmentsByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}
		var departments []models.Department
		db.Find(&departments)
		successResponse := helper.Response{
			Code:        http.StatusOK,
			Error:       false,
			Message:     "Departments retrieved successfully",
			Departments: departments,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllClientsByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
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
		db.Model(&models.Employee{}).Where("is_client = ?", true).
			Select("id", "first_name", "last_name", "full_name", "contact_number", "gender", "email", "username", "country", "is_active").
			Find(&clientEmployees)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Client  data retrieved successfully",
			"data":    clientEmployees,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAllLeaveRequestTypesByEmployee(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
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

		var leaveRequestTypes []models.LeaveRequestType
		db.Model(&models.LeaveRequestType{}).Offset(offset).Limit(perPage).Find(&leaveRequestTypes)

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

		var totalCount int64
		db.Model(&models.LeaveRequestType{}).Count(&totalCount)

		successResponse := map[string]interface{}{
			"code":                http.StatusOK,
			"error":               false,
			"message":             "Leave request types retrieved successfully",
			"leave_request_types": leaveRequestTypes,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
