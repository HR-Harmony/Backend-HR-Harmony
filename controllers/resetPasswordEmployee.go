package controllers

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/models"
)

// SendOTPForPasswordReset mengirimkan OTP ke email yang terdaftar untuk reset password
func SendOTPForPasswordReset(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Bind JSON request ke struct VerifyOTPRequest
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
		var employee models.Employee
		result := db.Where("email = ?", email).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Email not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Cek kapan terakhir kali OTP diminta
		var lastOTPRequest models.ResetPasswordOTP
		result = db.Where("employee_id = ? AND requested_at >= ?", employee.ID, time.Now().Add(-60*time.Second)).Order("requested_at desc").First(&lastOTPRequest)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusTooManyRequests, Message: "You can only request an OTP every 60 seconds"}
			return c.JSON(http.StatusTooManyRequests, errorResponse)
		}

		Fullname := employee.FirstName + " " + employee.LastName

		// Generate OTP
		otp := helper.GenerateOTP()

		// Menyimpan OTP ke dalam database
		expiredAt := time.Now().Add(15 * time.Minute)
		resetPasswordOTP := models.ResetPasswordOTP{
			EmployeeID:  employee.ID,
			OTP:         otp,
			ExpiredAt:   expiredAt,
			RequestedAt: time.Now(),
		}
		db.Create(&resetPasswordOTP)

		// Kirim OTP ke email
		err := helper.SendPasswordResetOTP(email, Fullname, otp, expiredAt)
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
// SendOTPForPasswordReset mengirimkan OTP ke email yang terdaftar untuk reset password
func SendOTPForPasswordReset(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Bind JSON request ke struct VerifyOTPRequest
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
		var employee models.Employee
		result := db.Where("email = ?", email).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Email not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		Fullname := employee.FirstName + " " + employee.LastName

		// Generate OTP
		otp := helper.GenerateOTP()

		// Menyimpan OTP ke dalam database
		expiredAt := time.Now().Add(15 * time.Minute)
		resetPasswordOTP := models.ResetPasswordOTP{
			EmployeeID: employee.ID,
			OTP:        otp,
			ExpiredAt:  expiredAt,
		}
		db.Create(&resetPasswordOTP)

		// Kirim OTP ke email
		err := helper.SendPasswordResetOTP(email, Fullname, otp, expiredAt)
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

// VerifyOTPForPasswordReset memverifikasi OTP yang dimasukkan oleh pengguna untuk reset password
func VerifyOTPForPasswordReset(db *gorm.DB) echo.HandlerFunc {
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
		var employee models.Employee
		result := db.Where("email = ?", email).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Email not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Cek apakah OTP sesuai dan masih berlaku
		var resetPasswordOTP models.ResetPasswordOTP
		result = db.Where("employee_id = ? AND otp = ? AND is_used = ? AND expired_at > ?", employee.ID, otp, false, time.Now()).First(&resetPasswordOTP)
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

// ResetPasswordWithOTP melakukan reset password dengan memeriksa OTP yang dimasukkan oleh pengguna
func ResetPasswordWithOTP(db *gorm.DB) echo.HandlerFunc {
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
		var employee models.Employee
		result := db.Where("email = ?", email).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Email not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Cek apakah OTP sesuai
		var resetPasswordOTP models.ResetPasswordOTP
		result = db.Where("employee_id = ? AND otp = ? AND expired_at > ?", employee.ID, otp, time.Now()).First(&resetPasswordOTP)
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
		employee.Password = string(hashedPassword)
		db.Save(&employee)

		// Menandai OTP sebagai digunakan
		resetPasswordOTP.IsUsed = true
		db.Save(&resetPasswordOTP)

		// Kirim notifikasi email bahwa password telah diubah
		if err := helper.SendPasswordChangedNotification(employee.Email, employee.FullName); err != nil {
			log.Println("Error sending password change notification email:", err) // Tambahkan log error di sini
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
