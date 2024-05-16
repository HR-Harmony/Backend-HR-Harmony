package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/models" // Sesuaikan dengan path model Employee
	"html/template"
	"net/http"
)

func VerifyEmail(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.QueryParam("token")

		var admin models.Admin
		result := db.Where("verification_token = ?", token).First(&admin)
		if result.Error != nil {
			return c.String(http.StatusUnauthorized, "Invalid verification token")
		}

		admin.IsVerified = true
		admin.VerificationToken = ""
		db.Save(&admin)

		tmpl, err := template.ParseFiles("helper/verification.html")
		if err != nil {
			return c.String(http.StatusInternalServerError, "Internal Server Error")
		}

		err = tmpl.Execute(c.Response().Writer, nil)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Internal Server Error")
		}

		return nil
	}
}
