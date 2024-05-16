package controllers

import (
	"errors"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"regexp"
)

func RegisterAdminHR(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		var admin models.Admin
		if err := c.Bind(&admin); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(admin.FirstName) < 3 || len(admin.LastName) < 3 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "First name and last name must be at least 3 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(admin.Username) < 5 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Username must be at least 5 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(admin.Password) < 8 || !helper.IsValidPassword(admin.Password) {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Password must be at least 8 characters and contain a combination of letters and numbers"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		match, _ := regexp.MatchString(emailPattern, admin.Email)
		if !match {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid email format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingAdmin models.Admin
		result := db.Where("username = ?", admin.Username).First(&existingAdmin)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Username already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to check username"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		result = db.Where("email = ?", admin.Email).First(&existingAdmin)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Email already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to check email"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to hash password"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		uniqueToken := helper.GenerateUniqueToken()
		admin.VerificationToken = uniqueToken

		admin.Password = string(hashedPassword)
		admin.IsAdminHR = true
		admin.Fullname = admin.FirstName + " " + admin.LastName
		db.Create(&admin)

		admin.Password = ""

		tokenString, err := middleware.GenerateToken(admin.Username, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to generate token"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		if err := helper.SendWelcomeEmail(admin.Email, admin.FirstName+" "+admin.LastName, uniqueToken); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send welcome email"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		response := map[string]interface{}{
			"code":    http.StatusOK,
			"message": "Admin HR registered successfully",
			"token":   tokenString,
			"id":      admin.ID,
		}

		return c.JSON(http.StatusOK, response)
	}
}
