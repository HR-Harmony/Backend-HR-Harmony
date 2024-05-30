package controllers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"hrsale/helper"
	"hrsale/middleware"
	"hrsale/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func calculateEarlyLeavingMinutes(earlyLeavingDuration string) int {
	earlyLeavingDur, _ := time.ParseDuration(earlyLeavingDuration)
	earlyLeavingMinutes := int(earlyLeavingDur.Minutes())
	return earlyLeavingMinutes
}

func AddManualAttendanceByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"})
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"})
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"})
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			return c.JSON(http.StatusNotFound, helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"})
		}

		if !adminUser.IsAdminHR {
			return c.JSON(http.StatusForbidden, helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"})
		}

		var attendance models.Attendance
		if err := c.Bind(&attendance); err != nil {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"})
		}

		if attendance.EmployeeID == 0 || attendance.AttendanceDate == "" || attendance.InTime == "" || attendance.OutTime == "" {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance data. All fields are required."})
		}

		var employee models.Employee
		result = db.First(&employee, attendance.EmployeeID)
		if result.Error != nil {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee ID not found"})
		}

		attendance.Username = employee.Username
		attendance.FullNameEmployee = employee.FirstName + " " + employee.LastName

		attendanceDate, err := time.Parse("2006-01-02", attendance.AttendanceDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance date format. Required format: yyyy-mm-dd"})
		}

		inTime, err := time.Parse("15:04:05", attendance.InTime)
		if err != nil {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid in_time format. Required format: HH:mm"})
		}

		outTime, err := time.Parse("15:04:05", attendance.OutTime)
		if err != nil {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid out_time format. Required format: HH:mm"})
		}

		shiftInTime, shiftOutTime, err := getShiftForDay(db, employee.ShiftID, attendanceDate.Weekday().String())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch shift data"})
		}

		lateDuration := calculateLate(shiftInTime, inTime.Format("15:04:05"))
		earlyLeavingDuration := calculateEarlyLeaving(shiftOutTime, outTime.Format("15:04:05"))

		workDuration := outTime.Sub(inTime)
		totalWorkHours := workDuration.Hours()
		totalWork := strconv.FormatFloat(totalWorkHours, 'f', 2, 64) + " hours"

		var hourlyRate float64
		if totalWorkHours < 0 {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid total work hours"})
		} else {
			hourlyRate = employee.HourlyRate
		}

		lateMinutes := calculateLateMinutes(lateDuration)
		earlyLeavingMinutes := calculateEarlyLeavingMinutes(earlyLeavingDuration)

		lateDeduction := (float64(lateMinutes) / 60) * hourlyRate
		earlyLeavingDeduction := (float64(earlyLeavingMinutes) / 60) * hourlyRate

		attendance.Status = "Present"
		attendance.Late = lateDuration
		attendance.LateMinutes = lateMinutes
		attendance.EarlyLeaving = earlyLeavingDuration
		attendance.EarlyLeavingMinutes = earlyLeavingMinutes
		attendance.TotalWork = totalWork

		db.Create(&attendance)

		successResponse := map[string]interface{}{
			"code":                    http.StatusCreated,
			"error":                   false,
			"message":                 "Attendance data added successfully",
			"data":                    attendance,
			"late_deduction":          lateDeduction,
			"early_leaving_deduction": earlyLeavingDeduction,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

/*
func AddManualAttendanceByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		var attendance models.Attendance
		if err := c.Bind(&attendance); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if attendance.EmployeeID == 0 || attendance.AttendanceDate == "" || attendance.InTime == "" || attendance.OutTime == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance data. All fields are required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var employee models.Employee
		result = db.First(&employee, attendance.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee ID not found"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		attendance.Username = employee.Username
		attendance.FullNameEmployee = employee.FirstName + " " + employee.LastName

		_, err = time.Parse("2006-01-02", attendance.AttendanceDate)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		inTime, err := time.Parse("15:04:00", attendance.InTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid in_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		outTime, err := time.Parse("15:04:00", attendance.OutTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid out_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		workDuration := outTime.Sub(inTime)

		totalWorkHours := workDuration.Hours()

		totalWork := strconv.FormatFloat(totalWorkHours, 'f', 2, 64) + " hours"

		attendance.TotalWork = totalWork

		attendance.Status = "Present"

		db.Create(&attendance)

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Attendance data added successfully",
			"data":    attendance,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}
*/

func GetAllAttendanceByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		date := c.QueryParam("date")
		employeeID := c.QueryParam("employee_id")
		searching := c.QueryParam("searching")

		query := db.Model(&models.Attendance{})

		if date != "" {
			query = query.Where("attendance_date = ?", date)
		}

		if employeeID != "" {
			query = query.Where("employee_id = ?", employeeID)
		}

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("full_name_employee ILIKE ? OR attendance_date ILIKE ?", searchPattern, searchPattern)
		}

		var totalCount int64
		query.Count(&totalCount)

		var attendance []models.Attendance
		query.Offset(offset).Limit(perPage).Find(&attendance)

		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Attendance data retrieved successfully",
			"data":       attendance,
			"pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAttendanceByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		attendanceID := c.Param("id")
		if attendanceID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Attendance ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var attendance models.Attendance
		result = db.First(&attendance, "id = ?", attendanceID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Attendance not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Attendance data retrieved successfully",
			"data":    attendance,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateAttendanceByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"})
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"})
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"})
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			return c.JSON(http.StatusNotFound, helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"})
		}

		if !adminUser.IsAdminHR {
			return c.JSON(http.StatusForbidden, helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"})
		}

		attendanceID := c.Param("id")
		if attendanceID == "" {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Attendance ID is missing"})
		}

		var attendance models.Attendance
		result = db.First(&attendance, "id = ?", attendanceID)
		if result.Error != nil {
			return c.JSON(http.StatusNotFound, helper.ErrorResponse{Code: http.StatusNotFound, Message: "Attendance not found"})
		}

		var updatedAttendance models.Attendance
		if err := c.Bind(&updatedAttendance); err != nil {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"})
		}

		var employee models.Employee
		if updatedAttendance.EmployeeID != 0 {
			result = db.First(&employee, "id = ?", updatedAttendance.EmployeeID)
			if result.Error != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid employee ID. Employee not found."})
			}
			attendance.EmployeeID = updatedAttendance.EmployeeID
			attendance.Username = employee.Username
			attendance.FullNameEmployee = employee.FirstName + " " + employee.LastName
		} else {
			result = db.First(&employee, "id = ?", attendance.EmployeeID)
			if result.Error != nil {
				return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"})
			}
		}

		var shouldRecalculate bool
		if updatedAttendance.AttendanceDate != "" {
			_, err := time.Parse("2006-01-02", updatedAttendance.AttendanceDate)
			if err != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance date format. Required format: yyyy-mm-dd"})
			}
			attendance.AttendanceDate = updatedAttendance.AttendanceDate
			shouldRecalculate = true
		}

		if updatedAttendance.InTime != "" {
			_, err := time.Parse("15:04:05", updatedAttendance.InTime)
			if err != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid in_time format. Required format: HH:mm:ss"})
			}
			attendance.InTime = updatedAttendance.InTime
			shouldRecalculate = true
		}

		if updatedAttendance.OutTime != "" {
			_, err := time.Parse("15:04:05", updatedAttendance.OutTime)
			if err != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid out_time format. Required format: HH:mm:ss"})
			}
			attendance.OutTime = updatedAttendance.OutTime
			shouldRecalculate = true
		}

		if shouldRecalculate {
			attendanceDate, err := time.Parse("2006-01-02", attendance.AttendanceDate)
			if err != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance date format. Required format: yyyy-mm-dd"})
			}
			inTime, err := time.Parse("15:04:05", attendance.InTime)
			if err != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid in_time format. Required format: HH:mm:ss"})
			}
			outTime, err := time.Parse("15:04:05", attendance.OutTime)
			if err != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid out_time format. Required format: HH:mm:ss"})
			}

			log.Printf("Fetching shift data for ShiftID: %d and Day: %s\n", employee.ShiftID, attendanceDate.Weekday().String())

			shiftInTime, shiftOutTime, err := getShiftForDay(db, employee.ShiftID, attendanceDate.Weekday().String())
			if err != nil {
				log.Printf("Failed to fetch shift data: %v\n", err)
				return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch shift data"})
			}

			lateDuration := calculateLate(shiftInTime, inTime.Format("15:04:05"))
			earlyLeavingDuration := calculateEarlyLeaving(shiftOutTime, outTime.Format("15:04:05"))

			workDuration := outTime.Sub(inTime)
			totalWorkHours := workDuration.Hours()
			totalWork := strconv.FormatFloat(totalWorkHours, 'f', 2, 64) + " hours"

			lateMinutes := calculateLateMinutes(lateDuration)
			earlyLeavingMinutes := calculateEarlyLeavingMinutes(earlyLeavingDuration)

			attendance.Status = "Present"
			attendance.Late = lateDuration
			attendance.LateMinutes = lateMinutes
			attendance.EarlyLeaving = earlyLeavingDuration
			attendance.EarlyLeavingMinutes = earlyLeavingMinutes
			attendance.TotalWork = totalWork
		}

		db.Save(&attendance)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Attendance data updated successfully",
			"data":    attendance,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func UpdateAttendanceByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		attendanceID := c.Param("id")
		if attendanceID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Attendance ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var attendance models.Attendance
		result = db.First(&attendance, "id = ?", attendanceID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Attendance not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedAttendance models.Attendance
		if err := c.Bind(&updatedAttendance); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedAttendance.EmployeeID != 0 {
			var employee models.Employee
			result = db.First(&employee, "id = ?", updatedAttendance.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid employee ID. Employee not found."}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			attendance.EmployeeID = updatedAttendance.EmployeeID
			attendance.Username = employee.Username
			attendance.FullNameEmployee = employee.FirstName + " " + employee.LastName
		}
		if updatedAttendance.AttendanceDate != "" {
			_, err := time.Parse("2006-01-02", updatedAttendance.AttendanceDate)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance date format. Required format: yyyy-mm-dd"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			attendance.AttendanceDate = updatedAttendance.AttendanceDate
		}
		if updatedAttendance.InTime != "" {
			attendance.InTime = updatedAttendance.InTime
		}
		if updatedAttendance.OutTime != "" {
			attendance.OutTime = updatedAttendance.OutTime
		}

		inTime, err := time.Parse("15:04:05", attendance.InTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid in_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		outTime, err := time.Parse("15:04:05", attendance.OutTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid out_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		workDuration := outTime.Sub(inTime)

		totalWorkHours := workDuration.Hours()

		totalWork := strconv.FormatFloat(totalWorkHours, 'f', 2, 64) + " hours"

		attendance.TotalWork = totalWork

		db.Save(&attendance)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Attendance data updated successfully",
			"data":    attendance,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func DeleteAttendanceByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		attendanceID := c.Param("id")
		if attendanceID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Attendance ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var attendance models.Attendance
		result = db.First(&attendance, "id = ?", attendanceID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Attendance not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&attendance)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Attendance data deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func CreateOvertimeRequestByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		var overtime models.OvertimeRequest
		if err := c.Bind(&overtime); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if overtime.EmployeeID == 0 || overtime.Date == "" || overtime.InTime == "" || overtime.OutTime == "" || overtime.Reason == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid Overtime Request. All fields are required."}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var employee models.Employee
		result = db.First(&employee, overtime.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee ID not found"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		overtime.Username = employee.Username
		overtime.FullNameEmployee = employee.FirstName + " " + employee.LastName

		_, err = time.Parse("2006-01-02", overtime.Date)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid Overtime Request date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		_, err = time.Parse("2006-01-02", overtime.Date)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		inTime, err := time.Parse("15:04", overtime.InTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid in_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		outTime, err := time.Parse("15:04", overtime.OutTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid out_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		workDuration := outTime.Sub(inTime)

		totalWorkHours := workDuration.Hours()

		totalWork := strconv.FormatFloat(totalWorkHours, 'f', 2, 64) + " hours"

		overtime.TotalWork = totalWork

		overtime.Status = "Pending"

		db.Create(&overtime)

		err = helper.SendOvertimeRequestNotification(employee.Email, overtime.FullNameEmployee, overtime.Date, overtime.InTime, overtime.OutTime, overtime.Reason)
		if err != nil {
			fmt.Println("Failed to send overtime request notification email:", err)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Overtime Request data added successfully",
			"data":    overtime,
		}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func GetAllOvertimeRequestsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10
		}

		offset := (page - 1) * perPage

		date := c.QueryParam("date")
		employeeID := c.QueryParam("employee_id")
		searching := c.QueryParam("searching")

		query := db.Model(&models.OvertimeRequest{})

		if date != "" {
			query = query.Where("date = ?", date)
		}

		if employeeID != "" {
			query = query.Where("employee_id = ?", employeeID)
		}

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("full_name_employee ILIKE ? OR date ILIKE ? OR status ILIKE ?", searchPattern, searchPattern, searchPattern)
		}

		var totalCount int64
		query.Count(&totalCount)

		var overtime []models.OvertimeRequest
		query.Offset(offset).Limit(perPage).Find(&overtime)

		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Overtime Request data retrieved successfully",
			"data":       overtime,
			"pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetOvertimeRequestByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		overtimeID := c.Param("id")
		if overtimeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Overtime ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var overtime models.OvertimeRequest
		result = db.First(&overtime, "id = ?", overtimeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Overtime Request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Attendance data retrieved successfully",
			"data":    overtime,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateOvertimeRequestByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		overtimeID := c.Param("id")
		if overtimeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Overtime ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var overtime models.OvertimeRequest
		result = db.First(&overtime, "id = ?", overtimeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Overtime Request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedOvertime models.OvertimeRequest
		if err := c.Bind(&updatedOvertime); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedOvertime.EmployeeID != 0 {
			var employee models.Employee
			result = db.First(&employee, "id = ?", updatedOvertime.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid employee ID. Employee not found."}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			overtime.EmployeeID = updatedOvertime.EmployeeID
			overtime.Username = employee.Username
			overtime.FullNameEmployee = employee.FirstName + " " + employee.LastName
		}
		if updatedOvertime.Date != "" {
			_, err := time.Parse("2006-01-02", updatedOvertime.Date)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid Overtime Request date format. Required format: yyyy-mm-dd"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			overtime.Date = updatedOvertime.Date
		}
		if updatedOvertime.InTime != "" {
			overtime.InTime = updatedOvertime.InTime
		}
		if updatedOvertime.OutTime != "" {
			overtime.OutTime = updatedOvertime.OutTime
		}

		if updatedOvertime.Reason != "" {
			overtime.Reason = updatedOvertime.Reason
		}

		if updatedOvertime.Status != "" {
			overtime.Status = updatedOvertime.Status
		}

		_, err = time.Parse("2006-01-02", overtime.Date)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance date format. Required format: yyyy-mm-dd"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		inTime, err := time.Parse("15:04", overtime.InTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid in_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		outTime, err := time.Parse("15:04", overtime.OutTime)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid out_time format. Required format: HH:mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		workDuration := outTime.Sub(inTime)

		totalWorkHours := workDuration.Hours()

		totalWork := strconv.FormatFloat(totalWorkHours, 'f', 2, 64) + " hours"

		overtime.TotalWork = totalWork

		db.Save(&overtime)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Attendance data updated successfully",
			"data":    overtime,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteOvertimeRequestByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		overtimeID := c.Param("id")
		if overtimeID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Overtime ID is missing"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var overtime models.OvertimeRequest
		result = db.First(&overtime, "id = ?", overtimeID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Overtime Request not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&overtime)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Overtime Request data deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func MarkAbsentEmployees(db *gorm.DB) {
	var employees []models.Employee
	db.Where("is_client = ? AND is_exit = ?", false, false).Find(&employees)

	today := time.Now().Format("2006-01-02")

	for _, employee := range employees {
		var existingAttendance models.Attendance
		result := db.Where("employee_id = ? AND attendance_date = ?", employee.ID, today).First(&existingAttendance)
		if result.Error != nil {
			shiftInTime, shiftOutTime, err := getShiftForDay(db, employee.ShiftID, time.Now().Weekday().String())
			if err != nil {
				log.Printf("Failed to fetch shift data for employee %s: %v\n", employee.Username, err)
				continue
			}

			// Parse shift times
			shiftIn, err := time.Parse("15:04:05", shiftInTime)
			if err != nil {
				log.Printf("Failed to parse shift in time for employee %s: %v\n", employee.Username, err)
				continue
			}
			shiftOut, err := time.Parse("15:04:05", shiftOutTime)
			if err != nil {
				log.Printf("Failed to parse shift out time for employee %s: %v\n", employee.Username, err)
				continue
			}

			// Calculate late minutes
			workDuration := shiftOut.Sub(shiftIn)
			lateMinutes := int(workDuration.Minutes())

			currentTime := time.Now()
			attendance := models.Attendance{
				EmployeeID:       employee.ID,
				Username:         employee.Username,
				FullNameEmployee: employee.FirstName + " " + employee.LastName,
				AttendanceDate:   today,
				InTime:           "-",
				OutTime:          "-",
				TotalWork:        "-",
				Status:           "Absent",
				LateMinutes:      lateMinutes,
				CreatedAt:        &currentTime,
			}
			db.Create(&attendance)
			log.Printf("Marked employee %s as absent on %s with late minutes %d\n", employee.Username, today, lateMinutes)
		}
	}
}

/*
func MarkAbsentEmployees(db *gorm.DB) {
	var employees []models.Employee
	db.Where("is_client = ? AND is_exit = ?", false, false).Find(&employees)

	today := time.Now().Format("2006-01-02")

	for _, employee := range employees {
		var existingAttendance models.Attendance
		result := db.Where("employee_id = ? AND attendance_date = ?", employee.ID, today).First(&existingAttendance)
		if result.Error != nil {
			currentTime := time.Now()
			attendance := models.Attendance{
				EmployeeID:       employee.ID,
				Username:         employee.Username,
				FullNameEmployee: employee.FirstName + " " + employee.LastName,
				AttendanceDate:   today,
				InTime:           "-",
				OutTime:          "-",
				TotalWork:        "-",
				Status:           "Absent",
				CreatedAt:        &currentTime,
			}
			db.Create(&attendance)
			log.Printf("Marked employee %s as absent on %s\n", employee.Username, today) // Add log here
		}
	}
}
*/
