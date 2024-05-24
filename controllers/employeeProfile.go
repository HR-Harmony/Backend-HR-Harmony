// controllers/employeeProfile.go

package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"strings"
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
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		employeeProfile := map[string]interface{}{
			"id":             employee.ID,
			"first_name":     employee.FirstName,
			"last_name":      employee.LastName,
			"full_name":      employee.FirstName + " " + employee.LastName,
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
			"basic_salary":   employee.BasicSalary,
			"hourly_rate":    employee.HourlyRate,
			"pay_slip_type":  employee.PaySlipType,
			"created_at":     employee.CreatedAt,
			"updated_at":     employee.UpdatedAt,
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Employee profile retrieved successfully",
			"profile": employeeProfile,
		})
	}
}
