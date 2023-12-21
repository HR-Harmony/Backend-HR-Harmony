package controllers

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
)

func SignInAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		var admin models.Admin
		if err := c.Bind(&admin); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validasi apakah username dan password telah diisi
		if admin.Username == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Username is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		if admin.Password == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Password is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Mengecek apakah username ada dalam database
		var existingAdmin models.Admin
		result := db.Where("username = ?", admin.Username).First(&existingAdmin)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid username"}
				return c.JSON(http.StatusUnauthorized, errorResponse)
			} else {
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to check username"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		// Membandingkan password yang dimasukkan dengan password yang di-hash
		err := bcrypt.CompareHashAndPassword([]byte(existingAdmin.Password), []byte(admin.Password))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid password"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		if !existingAdmin.IsVerified {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Account not verified. Please verify your account before logging in."}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Generate JWT token
		tokenString, err := middleware.GenerateToken(existingAdmin.Username, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to generate token"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Goroutine untuk mengirimkan email notifikasi
		go func(email, username string) {
			if err := helper.SendLoginNotification(email, username); err != nil {
				fmt.Println("Failed to send notification email:", err)
			}
		}(existingAdmin.Email, existingAdmin.FirstName+" "+existingAdmin.LastName)

		// Menyertakan ID pengguna dalam respons
		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "message": "Admin login successful", "token": tokenString, "id": existingAdmin.ID})
	}
}
