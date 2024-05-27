package controllers

import (
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/models"
	"net/http"
	"time"
)

// SendOTPForAdminPasswordReset mengirimkan OTP ke email yang terdaftar untuk reset password
func SendOTPForPasswordResetAdmin(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Bind JSON request ke struct SendOTPRequest
		type SendOTPRequest struct {
			Email string `json:"email"`
		}
		var req SendOTPRequest
		if err := c.Bind(&req); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		email := req.Email

		// Cek apakah email sudah terdaftar
		var admin models.Admin
		result := db.Where("email = ?", email).First(&admin)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Email not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		fullname := admin.FirstName + " " + admin.LastName

		// Generate OTP
		otp := helper.GenerateOTP()

		// Menyimpan OTP ke dalam database
		expiredAt := time.Now().Add(15 * time.Minute)
		resetPasswordOTP := models.AdminResetPasswordOTP{
			AdminID:   admin.ID,
			OTP:       otp,
			ExpiredAt: expiredAt,
		}
		db.Create(&resetPasswordOTP)

		// Kirim OTP ke email
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

// VerifyOTPForAdminPasswordReset memverifikasi OTP yang dimasukkan oleh pengguna untuk reset password
func VerifyOTPForPasswordResetAdmin(db *gorm.DB) echo.HandlerFunc {
	type VerifyOTPRequest struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}

	return func(c echo.Context) error {
		// Bind JSON request ke struct VerifyOTPRequest
		var req VerifyOTPRequest
		if err := c.Bind(&req); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		email := req.Email
		otp := req.OTP

		// Cek apakah email sudah terdaftar
		var admin models.Admin
		result := db.Where("email = ?", email).First(&admin)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Email not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Cek apakah OTP sesuai dan masih berlaku
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

// ResetAdminPasswordWithOTP melakukan reset password dengan memeriksa OTP yang dimasukkan oleh pengguna
func ResetPasswordWithOTPAdmin(db *gorm.DB) echo.HandlerFunc {
	type ResetPasswordRequest struct {
		Email              string `json:"email"`
		OTP                string `json:"otp"`
		NewPassword        string `json:"new_password"`
		ConfirmNewPassword string `json:"confirm_new_password"`
	}

	return func(c echo.Context) error {
		// Bind JSON request ke struct ResetPasswordRequest
		var req ResetPasswordRequest
		if err := c.Bind(&req); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		email := req.Email
		otp := req.OTP
		newPassword := req.NewPassword
		confirmNewPassword := req.ConfirmNewPassword

		// Cek apakah email sudah terdaftar
		var admin models.Admin
		result := db.Where("email = ?", email).First(&admin)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Email not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Cek apakah OTP sesuai
		var resetPasswordOTP models.AdminResetPasswordOTP
		result = db.Where("admin_id = ? AND otp = ? AND expired_at > ?", admin.ID, otp, time.Now()).First(&resetPasswordOTP)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid OTP"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Cek apakah password baru cocok dengan konfirmasi
		if newPassword != confirmNewPassword {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Passwords do not match"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if resetPasswordOTP.IsUsed {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "OTP has already been used"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Hash password baru
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to hash password"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Update password di database
		admin.Password = string(hashedPassword)
		db.Save(&admin)

		// Menandai OTP sebagai digunakan
		resetPasswordOTP.IsUsed = true
		db.Save(&resetPasswordOTP)

		// Kirim notifikasi email bahwa password telah diubah
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
