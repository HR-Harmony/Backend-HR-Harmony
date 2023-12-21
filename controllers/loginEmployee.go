// controllers/employeeLogin.go

package controllers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
)

// EmployeeLogin handles the login process for employees
func EmployeeLogin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		var employee models.Employee
		if err := c.Bind(&employee); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validasi apakah username dan password telah diisi
		if employee.Username == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Username is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		if employee.Password == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Password is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Mengecek apakah username ada dalam database
		var existingEmployee models.Employee
		result := db.Where("username = ?", employee.Username).First(&existingEmployee)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid username"}
				return c.JSON(http.StatusUnauthorized, errorResponse)
			} else {
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to check username"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		// Membandingkan password yang dimasukkan dengan password yang di-hash
		err := bcrypt.CompareHashAndPassword([]byte(existingEmployee.Password), []byte(employee.Password))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid password"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Generate JWT token
		tokenString, err := middleware.GenerateToken(existingEmployee.Username, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to generate token"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		go func(email, username string) {
			if err := helper.SendLoginNotification(email, username); err != nil {
				fmt.Println("Failed to send notification email:", err)
			}
		}(existingEmployee.Email, existingEmployee.FirstName+" "+existingEmployee.LastName)

		// Menyertakan ID karyawan dalam respons
		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "message": "Employee login successful", "token": tokenString, "id": existingEmployee.ID})
	}
}
