package controllers

import (
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/models"
	"log"
	"net/http"
	"time"
)

func SendOTPForPasswordResetAdmin(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		type SendOTPRequest struct {
			Email string `json:"email"`
		}
		var req SendOTPRequest
		if err := c.Bind(&req); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		email := req.Email

		var admin models.Admin
		result := db.Where("email = ?", email).First(&admin)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Email not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var lastOTPRequest models.AdminResetPasswordOTP
		result = db.Where("admin_id = ? AND requested_at >= ?", admin.ID, time.Now().Add(-60*time.Second)).Order("requested_at desc").First(&lastOTPRequest)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusTooManyRequests, Message: "You can only request an OTP every 60 seconds"}
			return c.JSON(http.StatusTooManyRequests, errorResponse)
		}

		fullname := admin.FirstName + " " + admin.LastName

		otp := helper.GenerateOTP()

		expiredAt := time.Now().Add(15 * time.Minute)
		resetPasswordOTP := models.AdminResetPasswordOTP{
			AdminID:     admin.ID,
			OTP:         otp,
			ExpiredAt:   expiredAt,
			RequestedAt: time.Now(),
		}
		db.Create(&resetPasswordOTP)

		err := helper.SendAdminPasswordResetOTP(email, fullname, otp, expiredAt)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send OTP"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "OTP sent to your email",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func SendOTPForPasswordResetAdmin(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		type SendOTPRequest struct {
			Email string `json:"email"`
		}
		var req SendOTPRequest
		if err := c.Bind(&req); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		email := req.Email

		var admin models.Admin
		result := db.Where("email = ?", email).First(&admin)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Email not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		fullname := admin.FirstName + " " + admin.LastName

		otp := helper.GenerateOTP()

		expiredAt := time.Now().Add(15 * time.Minute)
		resetPasswordOTP := models.AdminResetPasswordOTP{
			AdminID:   admin.ID,
			OTP:       otp,
			ExpiredAt: expiredAt,
		}
		db.Create(&resetPasswordOTP)

		err := helper.SendAdminPasswordResetOTP(email, fullname, otp, expiredAt)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send OTP"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "OTP sent to your email",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

/*
// versi 1 verify OTP
func VerifyOTPForPasswordResetAdmin(db *gorm.DB) echo.HandlerFunc {
	type VerifyOTPRequest struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}

	return func(c echo.Context) error {
		var req VerifyOTPRequest
		if err := c.Bind(&req); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		email := req.Email
		otp := req.OTP

		var admin models.Admin
		result := db.Where("email = ?", email).First(&admin)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Email not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var resetPasswordOTP models.AdminResetPasswordOTP
		result = db.Where("admin_id = ? AND otp = ? AND is_used = ? AND expired_at > ?", admin.ID, otp, false, time.Now()).First(&resetPasswordOTP)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid or expired OTP"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "OTP verified successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func VerifyOTPForPasswordResetAdmin(db *gorm.DB) echo.HandlerFunc {
	type VerifyOTPRequest struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}

	return func(c echo.Context) error {
		var req VerifyOTPRequest
		if err := c.Bind(&req); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		email := req.Email
		otp := req.OTP

		var admin models.Admin
		result := db.Where("email = ?", email).First(&admin)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Email not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Ambil OTP terakhir yang diminta oleh admin berdasarkan ID
		var lastRequestedOTP models.AdminResetPasswordOTP
		result = db.Where("admin_id = ?", admin.ID).Order("id desc").First(&lastRequestedOTP)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid or expired OTP"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Debug log
		log.Printf("Expected OTP: %s, Provided OTP: %s, IsUsed: %v, ExpiredAt: %v", lastRequestedOTP.OTP, otp, lastRequestedOTP.IsUsed, lastRequestedOTP.ExpiredAt)

		// Cek apakah OTP sesuai dan masih berlaku
		if lastRequestedOTP.OTP != otp || lastRequestedOTP.IsUsed || lastRequestedOTP.ExpiredAt.Before(time.Now()) {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid or expired OTP"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "OTP verified successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
// Reset password versi 1
func ResetPasswordWithOTPAdmin(db *gorm.DB) echo.HandlerFunc {
	type ResetPasswordRequest struct {
		Email              string `json:"email"`
		OTP                string `json:"otp"`
		NewPassword        string `json:"new_password"`
		ConfirmNewPassword string `json:"confirm_new_password"`
	}

	return func(c echo.Context) error {
		var req ResetPasswordRequest
		if err := c.Bind(&req); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		email := req.Email
		otp := req.OTP
		newPassword := req.NewPassword
		confirmNewPassword := req.ConfirmNewPassword

		var admin models.Admin
		result := db.Where("email = ?", email).First(&admin)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Email not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var resetPasswordOTP models.AdminResetPasswordOTP
		result = db.Where("admin_id = ? AND otp = ? AND expired_at > ?", admin.ID, otp, time.Now()).First(&resetPasswordOTP)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid OTP"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		if newPassword != confirmNewPassword {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Passwords do not match"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if resetPasswordOTP.IsUsed {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "OTP has already been used"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to hash password"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		admin.Password = string(hashedPassword)
		db.Save(&admin)

		resetPasswordOTP.IsUsed = true
		db.Save(&resetPasswordOTP)

		if err := helper.SendAdminPasswordChangedNotification(admin.Email, admin.Fullname); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send password change notification"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Password reset successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func ResetPasswordWithOTPAdmin(db *gorm.DB) echo.HandlerFunc {
	type ResetPasswordRequest struct {
		Email              string `json:"email"`
		OTP                string `json:"otp"`
		NewPassword        string `json:"new_password"`
		ConfirmNewPassword string `json:"confirm_new_password"`
	}

	return func(c echo.Context) error {
		var req ResetPasswordRequest
		if err := c.Bind(&req); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		email := req.Email
		otp := req.OTP
		newPassword := req.NewPassword
		confirmNewPassword := req.ConfirmNewPassword

		var admin models.Admin
		result := db.Where("email = ?", email).First(&admin)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Email not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Ambil OTP terakhir yang diminta oleh admin berdasarkan ID
		var lastRequestedOTP models.AdminResetPasswordOTP
		result = db.Where("admin_id = ?", admin.ID).Order("id desc").First(&lastRequestedOTP)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid or expired OTP"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Debug log
		log.Printf("Expected OTP: %s, Provided OTP: %s, IsUsed: %v, ExpiredAt: %v", lastRequestedOTP.OTP, otp, lastRequestedOTP.IsUsed, lastRequestedOTP.ExpiredAt)

		// Cek apakah OTP sesuai dan masih berlaku
		if lastRequestedOTP.OTP != otp || lastRequestedOTP.IsUsed || lastRequestedOTP.ExpiredAt.Before(time.Now()) {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid or expired OTP"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		if newPassword != confirmNewPassword {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Passwords do not match"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Hash password baru
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to hash password"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		admin.Password = string(hashedPassword)
		db.Save(&admin)

		lastRequestedOTP.IsUsed = true
		db.Save(&lastRequestedOTP)

		if err := helper.SendAdminPasswordChangedNotification(admin.Email, admin.Fullname); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send password change notification"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		successResponse := helper.Response{
			Code:    http.StatusOK,
			Error:   false,
			Message: "Password reset successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
