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

func getShiftForDay(db *gorm.DB, shiftID uint, day string) (string, string, error) {
	var shift models.Shift
	result := db.First(&shift, shiftID)
	if result.Error != nil {
		return "", "", result.Error
	}

	var inTime, outTime string
	switch day {
	case "Monday":
		inTime = shift.MondayInTime
		outTime = shift.MondayOutTime
	case "Tuesday":
		inTime = shift.TuesdayInTime
		outTime = shift.TuesdayOutTime
	case "Wednesday":
		inTime = shift.WednesdayInTime
		outTime = shift.WednesdayOutTime
	case "Thursday":
		inTime = shift.ThursdayInTime
		outTime = shift.ThursdayOutTime
	case "Friday":
		inTime = shift.FridayInTime
		outTime = shift.FridayOutTime
	case "Saturday":
		inTime = shift.SaturdayInTime
		outTime = shift.SaturdayOutTime
	case "Sunday":
		inTime = shift.SundayInTime
		outTime = shift.SundayOutTime
	default:
		return "", "", fmt.Errorf("invalid day: %s", day)
	}

	return inTime, outTime, nil
}

func calculateLate(shiftInTime string, actualInTime string) string {
	shiftIn, _ := time.Parse("15:04:05", shiftInTime)
	actualIn, _ := time.Parse("15:04:05", actualInTime)
	if actualIn.After(shiftIn) {
		lateDuration := actualIn.Sub(shiftIn).Round(time.Minute)
		return lateDuration.String()
	}
	return "0s"
}

func calculateLateMinutes(late string) int {
	duration, err := time.ParseDuration(late)
	if err != nil {
		fmt.Println("Error parsing late duration:", err)
		return 0
	}
	return int(duration.Minutes())
}

func calculateEarlyLeaving(shiftOutTime string, actualOutTime string) string {
	shiftOut, _ := time.Parse("15:04:05", shiftOutTime)
	actualOut, _ := time.Parse("15:04:05", actualOutTime)
	if actualOut.Before(shiftOut) {
		earlyLeavingDuration := shiftOut.Sub(actualOut).Round(time.Minute)
		return earlyLeavingDuration.String()
	}
	return "0s"
}

func EmployeeCheckIn(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Load Jakarta timezone
		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to load timezone"})
		}

		today := time.Now().In(loc).Format("2006-01-02")
		var existingAttendance models.Attendance
		result = db.Where("employee_id = ? AND attendance_date = ?", employee.ID, today).First(&existingAttendance)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee has already checked in for today"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		employee.FullName = employee.FirstName + " " + employee.LastName

		currentTime := time.Now().In(loc)
		shiftInTime, _, err := getShiftForDay(db, employee.ShiftID, currentTime.Weekday().String())
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch shift data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		lateDuration := calculateLate(shiftInTime, currentTime.Format("15:04:05"))
		lateMinutes := calculateLateMinutes(lateDuration)

		attendance := models.Attendance{
			EmployeeID:       employee.ID,
			Username:         employee.Username,
			FullNameEmployee: employee.FirstName + " " + employee.LastName,
			AttendanceDate:   today,
			InTime:           currentTime.Format("15:04:05"),
			Status:           "Present",
			Late:             lateDuration,
			LateMinutes:      lateMinutes,
			CreatedAt:        &currentTime,
		}
		db.Create(&attendance)

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
			"late":    attendance.Late,
		})
	}
}

/*
Employee checkin yang waktunya belum memakai format waktu indonesia

func EmployeeCheckIn(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		today := time.Now().Format("2006-01-02")
		var existingAttendance models.Attendance
		result = db.Where("employee_id = ? AND attendance_date = ?", employee.ID, today).First(&existingAttendance)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee has already checked in for today"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		employee.FullName = employee.FirstName + " " + employee.LastName

		currentTime := time.Now()
		shiftInTime, _, err := getShiftForDay(db, employee.ShiftID, currentTime.Weekday().String())
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch shift data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		lateDuration := calculateLate(shiftInTime, currentTime.Format("15:04:05"))
		lateMinutes := calculateLateMinutes(lateDuration)

		attendance := models.Attendance{
			EmployeeID:       employee.ID,
			Username:         employee.Username,
			FullNameEmployee: employee.FirstName + " " + employee.LastName,
			AttendanceDate:   today,
			InTime:           currentTime.Format("15:04:05"),
			Status:           "Present",
			Late:             lateDuration,
			LateMinutes:      lateMinutes,
			CreatedAt:        &currentTime,
		}
		db.Create(&attendance)

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
			"late":    attendance.Late,
		})
	}
}
*/

/*
func EmployeeCheckIn(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		today := time.Now().Format("2006-01-02")
		var existingAttendance models.Attendance
		result = db.Where("employee_id = ? AND attendance_date = ?", employee.ID, today).First(&existingAttendance)
		if result.Error == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee has already checked in for today"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		employee.FullName = employee.FirstName + " " + employee.LastName

		currentTime := time.Now()
		attendance := models.Attendance{
			EmployeeID:       employee.ID,
			Username:         employee.Username,
			FullNameEmployee: employee.FirstName + " " + employee.LastName,
			AttendanceDate:   today,
			InTime:           currentTime.Format("15:04:05"),
			Status:           "Present",
			CreatedAt:        &currentTime,
		}
		db.Create(&attendance)

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
*/

func EmployeeCheckOut(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Load Jakarta timezone
		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to load timezone"})
		}

		today := time.Now().In(loc).Format("2006-01-02")
		var existingAttendance models.Attendance
		result = db.Where("employee_id = ? AND attendance_date = ?", employee.ID, today).First(&existingAttendance)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee has not checked in for today"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if existingAttendance.Status == "Absent" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee is absent for today"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if existingAttendance.OutTime != "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee has already checked out for today"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		currentTime := time.Now().In(loc)
		existingAttendance.OutTime = currentTime.Format("15:04:05")
		inTime, _ := time.Parse("15:04:05", existingAttendance.InTime)
		outTime, _ := time.Parse("15:04:05", existingAttendance.OutTime)
		totalWork := outTime.Sub(inTime).Round(time.Minute)
		existingAttendance.TotalWork = totalWork.String()

		_, shiftOutTime, err := getShiftForDay(db, employee.ShiftID, currentTime.Weekday().String())
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch shift data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		earlyLeavingDuration := calculateEarlyLeaving(shiftOutTime, currentTime.Format("15:04:05"))
		existingAttendance.EarlyLeaving = earlyLeavingDuration
		earlyLeavingMinutes := calculateEarlyLeavingMinutes(earlyLeavingDuration)
		existingAttendance.EarlyLeavingMinutes = earlyLeavingMinutes

		db.Save(&existingAttendance)

		err = helper.SendAttendanceCheckoutNotification(employee.Email, employee.FirstName+" "+employee.LastName, existingAttendance.OutTime, existingAttendance.TotalWork)
		if err != nil {
			// Handle error
			fmt.Println("Failed to send checkout notification:", err)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":          http.StatusOK,
			"error":         false,
			"message":       "Employee check-out successful",
			"time":          existingAttendance.OutTime,
			"total_work":    existingAttendance.TotalWork,
			"early_leaving": existingAttendance.EarlyLeaving,
		})
	}
}

/*
Checkout yang belum memakai waktu jakarta

func EmployeeCheckOut(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		today := time.Now().Format("2006-01-02")
		var existingAttendance models.Attendance
		result = db.Where("employee_id = ? AND attendance_date = ?", employee.ID, today).First(&existingAttendance)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee has not checked in for today"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if existingAttendance.Status == "Absent" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee is absent for today"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if existingAttendance.OutTime != "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee has already checked out for today"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		currentTime := time.Now()
		existingAttendance.OutTime = currentTime.Format("15:04:05")
		inTime, _ := time.Parse("15:04:05", existingAttendance.InTime)
		outTime, _ := time.Parse("15:04:05", existingAttendance.OutTime)
		totalWork := outTime.Sub(inTime).Round(time.Minute)
		existingAttendance.TotalWork = totalWork.String()

		_, shiftOutTime, err := getShiftForDay(db, employee.ShiftID, currentTime.Weekday().String())
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch shift data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		earlyLeavingDuration := calculateEarlyLeaving(shiftOutTime, currentTime.Format("15:04:05"))
		existingAttendance.EarlyLeaving = earlyLeavingDuration
		earlyLeavingMinutes := calculateEarlyLeavingMinutes(earlyLeavingDuration)
		existingAttendance.EarlyLeavingMinutes = earlyLeavingMinutes

		db.Save(&existingAttendance)

		err = helper.SendAttendanceCheckoutNotification(employee.Email, employee.FirstName+" "+employee.LastName, existingAttendance.OutTime, existingAttendance.TotalWork)
		if err != nil {
			// Handle error
			fmt.Println("Failed to send checkout notification:", err)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":          http.StatusOK,
			"error":         false,
			"message":       "Employee check-out successful",
			"time":          existingAttendance.OutTime,
			"total_work":    existingAttendance.TotalWork,
			"early_leaving": existingAttendance.EarlyLeaving,
		})
	}
}
*/

/*
func EmployeeCheckOut(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		today := time.Now().Format("2006-01-02")
		var existingAttendance models.Attendance
		result = db.Where("employee_id = ? AND attendance_date = ?", employee.ID, today).First(&existingAttendance)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee has not checked in for today"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if existingAttendance.Status == "Absent" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee is absent for today"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if already checked out
		if existingAttendance.OutTime != "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Employee has already checked out for today"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		currentTime := time.Now()
		existingAttendance.OutTime = currentTime.Format("15:04:05")
		inTime, _ := time.Parse("15:04:05", existingAttendance.InTime)
		outTime, _ := time.Parse("15:04:05", existingAttendance.OutTime)
		totalWork := outTime.Sub(inTime).Round(time.Minute)
		existingAttendance.TotalWork = totalWork.String()

		db.Save(&existingAttendance)

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
*/

func EmployeeAttendance(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve employee details
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Pagination parameters
		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10 // Default per page
		}

		// Calculate offset and limit for pagination
		offset := (page - 1) * perPage

		// Query parameters for searching
		searching := c.QueryParam("searching")

		// Build the query
		query := db.Model(&models.Attendance{}).Where("employee_id = ?", employee.ID)
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where(
				"LOWER(full_name_employee) LIKE ? OR LOWER(attendance_date) LIKE ? OR LOWER(in_time) LIKE ? OR LOWER(out_time) LIKE ? OR LOWER(status) LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
			)
		}

		// Retrieve attendances for the employee with pagination
		var attendances []models.Attendance
		result = query.Order("id DESC").Offset(offset).Limit(perPage).Find(&attendances)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch attendance data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Get total count of attendances for the employee
		var totalCount int64
		query.Count(&totalCount)

		// Batch fetch employee full names
		var employeeIDs []uint
		employeeMap := make(map[uint]string)
		for _, att := range attendances {
			employeeIDs = append(employeeIDs, att.EmployeeID)
			employeeMap[att.EmployeeID] = ""
		}

		var employees []models.Employee
		db.Where("id IN (?)", employeeIDs).Find(&employees)

		for _, emp := range employees {
			employeeMap[emp.ID] = emp.FullName
		}

		// Assign full names to attendances and update database
		tx := db.Begin()
		for i := range attendances {
			if fullName, ok := employeeMap[attendances[i].EmployeeID]; ok {
				attendances[i].FullNameEmployee = fullName
			}
			if err := tx.Save(&attendances[i]).Error; err != nil {
				tx.Rollback()
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Error saving attendance data"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}
		tx.Commit()

		// Provide success response
		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Employee attendance data retrieved successfully",
			"attendance": attendances,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
func EmployeeAttendance(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		// Retrieve employee details
		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Pagination parameters
		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.QueryParam("per_page"))
		if err != nil || perPage <= 0 {
			perPage = 10 // Default per page
		}

		// Calculate offset and limit for pagination
		offset := (page - 1) * perPage

		// Query parameters for searching
		searching := c.QueryParam("searching")

		// Build the query
		query := db.Model(&models.Attendance{}).Where("employee_id = ?", employee.ID)
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where(
				"LOWER(full_name_employee) LIKE ? OR LOWER(attendance_date) LIKE ? OR LOWER(in_time) LIKE ? OR LOWER(out_time) LIKE ? OR LOWER(status) LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
			)
		}

		// Retrieve attendances for the employee with pagination
		var attendances []models.Attendance
		result = query.Preload("Employee").Order("id DESC").Offset(offset).Limit(perPage).Find(&attendances)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch attendance data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Get total count of attendances for the employee
		var totalCount int64
		query.Count(&totalCount)

		// Provide success response
		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Employee attendance data retrieved successfully",
			"attendance": attendances,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

func EmployeeAttendanceByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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

		var employee models.Employee
		result := db.Where("username = ?", username).First(&employee)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch employee data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		attendanceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid attendance ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var attendance models.Attendance
		result = db.Preload("Employee").Where("id = ?", attendanceID).First(&attendance)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Attendance ID not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch attendance data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		if attendance.EmployeeID != employee.ID {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Attendance does not belong to the employee"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Attendance data retrieved successfully",
			"attendance": attendance,
		})
	}
}
