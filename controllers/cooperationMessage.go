package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/models"
	"net/http"
	"time"
)

const adminEmail = "hriscloud@gmail.com"

func CreateCooperationMessage(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var cooperationMessage models.CooperationMessage

		if err := c.Bind(&cooperationMessage); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(cooperationMessage.FirstName) < 3 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "first name harus minimal 3 huruf"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if !helper.IsValidEmail(cooperationMessage.Email) {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Format email tidak valid"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if !helper.IsValidPhoneNumber(cooperationMessage.PhoneNumber) {
			if !helper.ContainsOnlyDigits(cooperationMessage.PhoneNumber) {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Phone number harus mengandung angka semua"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			} else if len(cooperationMessage.PhoneNumber) < 10 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Phone number kurang dari 10 digit"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		if len(cooperationMessage.Message) < 10 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Message harus minimal 10 huruf"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		cooperationMessage.CreatedAt = time.Now()

		db.Create(&cooperationMessage)

		adminEmailSubject := "New Cooperation Message"
		adminEmailBody := helper.GetCooperationEmailBody(cooperationMessage)
		if err := helper.SendEmailToUser(adminEmail, adminEmailSubject, adminEmailBody); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send email to admin"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		userEmailSubject := "Your Cooperation Message"
		userEmailBody := helper.GetUserCooperationEmailBody(cooperationMessage)
		if err := helper.SendEmailToUser(cooperationMessage.Email, userEmailSubject, userEmailBody); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send email to user"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{"code": http.StatusCreated, "error": false, "message": "Cooperation Message berhasil dikirim"})
	}
}
