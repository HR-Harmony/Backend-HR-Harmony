package controllers

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func EmployeeCheckIn(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the token from the request header
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

		// Verify the token and extract the username
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Check if the employee has already checked in for the day
		today := time.Now().Format("2006-01-02")
		var existingAttendance models.Attendance
		result = db.Where("employee_id = ? AND attendance_date = ?", employee.ID, today).First(&existingAttendance)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee has already checked in for today"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Record the check-in time
		currentTime := time.Now()
		attendance := models.Attendance{
			EmployeeID:     employee.ID,
			Username:       employee.Username,
			AttendanceDate: today,
			InTime:         currentTime.Format("15:04:05"),
			CreatedAt:      &currentTime,
		}
		db.Create(&attendance)

		// Kirim notifikasi email ke karyawan
		err = helper.SendAttendanceCheckinNotification(employee.Email, employee.FirstName+" "+employee.LastName, attendance.InTime)
		if err != nil {
			// Handle error
			fmt.Println("Failed to send check-in notification:", err)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Employee check-in successful",
			"time":    attendance.InTime,
		})
	}
}

func EmployeeCheckOut(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the token from the request header
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

		// Verify the token and extract the username
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Check if the employee has checked in for the day
		today := time.Now().Format("2006-01-02")
		var existingAttendance models.Attendance
		result = db.Where("employee_id = ? AND attendance_date = ?", employee.ID, today).First(&existingAttendance)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee has not checked in for today"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Update the check-out time and calculate total work
		currentTime := time.Now()
		existingAttendance.OutTime = currentTime.Format("15:04:05")
		inTime, _ := time.Parse("15:04:05", existingAttendance.InTime)
		outTime, _ := time.Parse("15:04:05", existingAttendance.OutTime)
		totalWork := outTime.Sub(inTime).Round(time.Minute)
		existingAttendance.TotalWork = totalWork.String()

		// Update the attendance record
		db.Save(&existingAttendance)

		// Kirim notifikasi email ke karyawan
		err = helper.SendAttendanceCheckoutNotification(employee.Email, employee.FirstName+" "+employee.LastName, existingAttendance.OutTime, existingAttendance.TotalWork)
		if err != nil {
			// Handle error
			fmt.Println("Failed to send checkout notification:", err)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Employee check-out successful",
			"time":       existingAttendance.OutTime,
			"total_work": existingAttendance.TotalWork,
		})
	}
}

func EmployeeAttendance(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the token from the request header
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

		// Verify the token and extract the username
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Retrieve attendance records for the employee
		var attendances []models.Attendance
		result = db.Where("employee_id = ?", employee.ID).Find(&attendances)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch attendance data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Return the attendance records
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Employee attendance data retrieved successfully",
			"attendance": attendances,
		})
	}
}

// GetAttendanceByIDAndEmployeeID mengambil data attendance milik karyawan berdasarkan ID attendance dan ID karyawan
func EmployeeAttendanceByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract and verify the JWT token
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

		// Retrieve the employee based on the username
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Extract attendanceID from the request parameters
		attendanceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Retrieve attendance data for the specified ID
		var attendance models.Attendance
		result = db.Where("id = ?", attendanceID).First(&attendance)
		if result.Error != nil {
			// Check if the attendance ID is not found
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Attendance ID not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch attendance data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Check if the retrieved attendance belongs to the employee
		if attendance.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Attendance does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Return the attendance data
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Attendance data retrieved successfully",
			"attendance": attendance,
		})
	}
}
