package controllers

import (
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

func GetAllEmployeesPayrollInfo(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
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

		searching := c.QueryParam("searching")

		query := db.Model(&models.Employee{}).Where("is_client = ? AND is_exit = ?", false, false)
		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where("LOWER(full_name) LIKE ?", searchPattern)
		}

		var employees []models.Employee
		result = query.Order("id DESC").Offset(offset).Limit(perPage).Find(&employees)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to retrieve employees"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var payrollInfoList []map[string]interface{}
		for _, employee := range employees {
			payrollInfo := map[string]interface{}{
				"payroll_id":   employee.PayrollID,
				"username":     employee.Username,
				"full_name":    employee.FullName,
				"employee_id":  employee.ID,
				"payslip_type": employee.PaySlipType,
				"basic_salary": employee.BasicSalary,
				"hourly_rate":  employee.HourlyRate,
				"paid_status":  employee.PaidStatus,
			}
			payrollInfoList = append(payrollInfoList, payrollInfo)
		}

		var totalCount int64
		query.Count(&totalCount)

		successResponse := map[string]interface{}{
			"Code":        http.StatusOK,
			"Error":       false,
			"Message":     "Employee payroll information retrieved successfully",
			"PayrollInfo": payrollInfoList,
			"Pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

// UpdatePaidStatusByPayrollID merupakan handler untuk memperbarui status pembayaran gaji berdasarkan ID payroll
func UpdatePaidStatusByPayrollID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Mendapatkan token dari header Authorization
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Memeriksa format token
		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		// Memverifikasi token
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mendapatkan data admin berdasarkan username
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Memeriksa apakah admin memiliki akses sebagai admin HR
		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Mendapatkan ID payroll dari URL parameter
		payrollID := c.Param("payroll_id")

		// Mendapatkan data employee berdasarkan payrollID
		var employee models.Employee
		if err := db.Preload("PayrollInfo").Where("payroll_id = ?", payrollID).First(&employee).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Mendapatkan bulan dan tahun saat ini
		currentMonth := time.Now().Format("2006-01")

		// Mendapatkan semua data kehadiran employee pada bulan dan tahun saat ini
		var attendances []models.Attendance
		result = db.Where("employee_id = ? AND attendance_date LIKE ?", employee.ID, currentMonth+"%").Find(&attendances)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch attendances"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Menghitung total menit keterlambatan dan total menit early leaving dari semua kehadiran pada bulan ini
		totalLateMinutes := 0
		totalEarlyLeavingMinutes := 0
		for _, attendance := range attendances {
			totalLateMinutes += attendance.LateMinutes
			totalEarlyLeavingMinutes += attendance.EarlyLeavingMinutes
		}

		// Menghitung potongan gaji berdasarkan total menit keterlambatan dan total menit early leaving
		lateDeduction := (employee.HourlyRate / 60) * float64(totalLateMinutes)
		earlyLeavingDeduction := (employee.HourlyRate / 60) * float64(totalEarlyLeavingMinutes)

		// Calculate total minutes of accepted overtime requests
		var acceptedOvertimes []models.OvertimeRequest
		result = db.Where("employee_id = ? AND status = ?", employee.ID, "Accepted").Find(&acceptedOvertimes)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch overtime requests"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		totalOvertimeMinutes := 0
		for _, overtime := range acceptedOvertimes {
			totalOvertimeMinutes += overtime.TotalMinutes
		}

		// Calculate overtime pay
		overtimePay := (employee.HourlyRate / 60) * float64(totalOvertimeMinutes)

		// Calculate loan deductions for approved loans
		var approvedLoans []models.RequestLoan
		result = db.Where("employee_id = ? AND status = ?", employee.ID, "Approved").Find(&approvedLoans)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch loan requests"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		totalLoanDeduction := 0
		for _, loan := range approvedLoans {
			if loan.Remaining > 0 {
				if loan.Remaining >= loan.MonthlyInstallmentAmt {
					totalLoanDeduction += loan.MonthlyInstallmentAmt
					loan.Remaining -= loan.MonthlyInstallmentAmt
				} else {
					totalLoanDeduction += loan.Remaining
					loan.Remaining = 0
				}
				db.Save(&loan)
			}
		}

		// Calculate final salary after all deductions and additions
		finalSalary := employee.BasicSalary - lateDeduction - earlyLeavingDeduction + overtimePay - float64(totalLoanDeduction)

		// Update employee's paid status and create payroll info
		employee.PaidStatus = true
		db.Save(&employee)

		// Membuat catatan pembayaran gaji
		payrollInfo := models.PayrollInfo{
			EmployeeID:       employee.ID,
			BasicSalary:      finalSalary,
			PayslipType:      employee.PaySlipType,
			PaidStatus:       employee.PaidStatus,
			FullNameEmployee: employee.FirstName + " " + employee.LastName,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		db.Create(&payrollInfo)

		// Mereset total menit keterlambatan, total menit early leaving, dan total menit lembur untuk bulan berikutnya
		db.Model(&models.Attendance{}).Where("employee_id = ?", employee.ID).Updates(map[string]interface{}{"late_minutes": 0, "early_leaving_minutes": 0})
		db.Model(&models.OvertimeRequest{}).Where("employee_id = ? AND status = ?", employee.ID, "Accepted").Updates(map[string]interface{}{"total_minutes": 0})

		/*
			db.Model(&models.OvertimeRequest{}).Where("employee_id = ?", employee.ID).Updates(map[string]interface{}{"total_minutes": 0})
		*/

		// Mengirim notifikasi email tentang pembayaran gaji
		go func(email, fullName string, finalSalary float64) {
			if err := helper.SendSalaryTransferNotification(email, fullName, finalSalary); err != nil {
				fmt.Println("Failed to send salary transfer notification email:", err)
			}
		}(employee.Email, employee.FirstName+" "+employee.LastName, finalSalary)

		// Membuat response sukses
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Paid status updated successfully",
			"employee": map[string]interface{}{
				"id":             employee.ID,
				"payroll_id":     employee.PayrollID,
				"first_name":     employee.FirstName,
				"last_name":      employee.LastName,
				"full_name":      employee.FullName,
				"contact_number": employee.ContactNumber,
				"gender":         employee.Gender,
				"email":          employee.Email,
				"username":       employee.Username,
				"shift_id":       employee.ShiftID,
				"shift":          employee.Shift,
				"role_id":        employee.RoleID,
				"role":           employee.Role,
				"department_id":  employee.DepartmentID,
				"department":     employee.Department,
				"designation_id": employee.DesignationID,
				"designation":    employee.Designation,
				"basic_salary":   employee.BasicSalary,
				"hourly_rate":    employee.HourlyRate,
				"pay_slip_type":  employee.PaySlipType,
				"is_active":      employee.IsActive,
				"paid_status":    employee.PaidStatus,
				"final_salary":   finalSalary,
				"created_at":     employee.CreatedAt,
				"updated_at":     employee.UpdatedAt,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

/*
// UpdatePaidStatusByPayrollID merupakan handler untuk memperbarui status pembayaran gaji berdasarkan ID payroll V1 Final
func UpdatePaidStatusByPayrollID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Mendapatkan token dari header Authorization
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Memeriksa format token
		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		// Memverifikasi token
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mendapatkan data admin berdasarkan username
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Memeriksa apakah admin memiliki akses sebagai admin HR
		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Mendapatkan ID payroll dari URL parameter
		payrollID := c.Param("payroll_id")

		// Mendapatkan data employee berdasarkan payrollID
		var employee models.Employee
		if err := db.Preload("PayrollInfo").Where("payroll_id = ?", payrollID).First(&employee).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Mendapatkan bulan dan tahun saat ini
		currentMonth := time.Now().Format("2006-01")

		// Mendapatkan semua data kehadiran employee pada bulan dan tahun saat ini
		var attendances []models.Attendance
		result = db.Where("employee_id = ? AND attendance_date LIKE ?", employee.ID, currentMonth+"%").Find(&attendances)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch attendances"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Menghitung total menit keterlambatan dan total menit early leaving dari semua kehadiran pada bulan ini
		totalLateMinutes := 0
		totalEarlyLeavingMinutes := 0
		for _, attendance := range attendances {
			totalLateMinutes += attendance.LateMinutes
			totalEarlyLeavingMinutes += attendance.EarlyLeavingMinutes
		}

		// Menghitung potongan gaji berdasarkan total menit keterlambatan dan total menit early leaving
		lateDeduction := (employee.HourlyRate / 60) * float64(totalLateMinutes)
		earlyLeavingDeduction := (employee.HourlyRate / 60) * float64(totalEarlyLeavingMinutes)

		// Calculate total minutes of accepted overtime requests
		var acceptedOvertimes []models.OvertimeRequest
		result = db.Where("employee_id = ? AND status = ?", employee.ID, "Accepted").Find(&acceptedOvertimes)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch overtime requests"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		totalOvertimeMinutes := 0
		for _, overtime := range acceptedOvertimes {
			totalOvertimeMinutes += overtime.TotalMinutes
		}

		// Calculate overtime pay
		overtimePay := (employee.HourlyRate / 60) * float64(totalOvertimeMinutes)
		finalSalary := employee.BasicSalary - lateDeduction - earlyLeavingDeduction + overtimePay

		// Update employee's paid status and create payroll info
		employee.PaidStatus = true
		db.Save(&employee)

		// Membuat catatan pembayaran gaji
		payrollInfo := models.PayrollInfo{
			EmployeeID:       employee.ID,
			BasicSalary:      finalSalary,
			PayslipType:      employee.PaySlipType,
			PaidStatus:       employee.PaidStatus,
			FullNameEmployee: employee.FirstName + " " + employee.LastName,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		db.Create(&payrollInfo)

		// Mereset total menit keterlambatan, total menit early leaving, dan total menit lembur untuk bulan berikutnya
		db.Model(&models.Attendance{}).Where("employee_id = ?", employee.ID).Updates(map[string]interface{}{"late_minutes": 0, "early_leaving_minutes": 0})
		db.Model(&models.OvertimeRequest{}).Where("employee_id = ? AND status = ?", employee.ID, "Accepted").Updates(map[string]interface{}{"total_minutes": 0})

		// Mengirim notifikasi email tentang pembayaran gaji
		go func(email, fullName string, finalSalary float64) {
			if err := helper.SendSalaryTransferNotification(email, fullName, finalSalary); err != nil {
				fmt.Println("Failed to send salary transfer notification email:", err)
			}
		}(employee.Email, employee.FirstName+" "+employee.LastName, finalSalary)

		// Membuat response sukses
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Paid status updated successfully",
			"employee": map[string]interface{}{
				"id":             employee.ID,
				"payroll_id":     employee.PayrollID,
				"first_name":     employee.FirstName,
				"last_name":      employee.LastName,
				"full_name":      employee.FullName,
				"contact_number": employee.ContactNumber,
				"gender":         employee.Gender,
				"email":          employee.Email,
				"username":       employee.Username,
				"shift_id":       employee.ShiftID,
				"shift":          employee.Shift,
				"role_id":        employee.RoleID,
				"role":           employee.Role,
				"department_id":  employee.DepartmentID,
				"department":     employee.Department,
				"designation_id": employee.DesignationID,
				"designation":    employee.Designation,
				"basic_salary":   employee.BasicSalary,
				"hourly_rate":    employee.HourlyRate,
				"pay_slip_type":  employee.PaySlipType,
				"is_active":      employee.IsActive,
				"paid_status":    employee.PaidStatus,
				"final_salary":   finalSalary,
				"created_at":     employee.CreatedAt,
				"updated_at":     employee.UpdatedAt,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

/*
// UpdatePaidStatusByPayrollID merupakan handler untuk memperbarui status pembayaran gaji berdasarkan ID payroll dilengkapi dengan pemotongan gaji berdasarkan attandances
func UpdatePaidStatusByPayrollID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Mendapatkan token dari header Authorization
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Memeriksa format token
		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		// Memverifikasi token
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mendapatkan data admin berdasarkan username
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Memeriksa apakah admin memiliki akses sebagai admin HR
		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Mendapatkan ID payroll dari URL parameter
		payrollID := c.Param("payroll_id")

		// Mendapatkan data employee berdasarkan payrollID
		var employee models.Employee
		if err := db.Preload("PayrollInfo").Where("payroll_id = ?", payrollID).First(&employee).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Mendapatkan bulan dan tahun saat ini
		currentMonth := time.Now().Format("2006-01")

		// Mendapatkan semua data kehadiran employee pada bulan dan tahun saat ini
		var attendances []models.Attendance
		result = db.Where("employee_id = ? AND attendance_date LIKE ?", employee.ID, currentMonth+"%").Find(&attendances)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch attendances"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Menghitung total menit keterlambatan dan total menit early leaving dari semua kehadiran pada bulan ini
		totalLateMinutes := 0
		totalEarlyLeavingMinutes := 0
		for _, attendance := range attendances {
			totalLateMinutes += attendance.LateMinutes
			totalEarlyLeavingMinutes += attendance.EarlyLeavingMinutes
		}

		// Menghitung potongan gaji berdasarkan total menit keterlambatan dan total menit early leaving
		lateDeduction := (employee.HourlyRate / 60) * float64(totalLateMinutes)
		earlyLeavingDeduction := (employee.HourlyRate / 60) * float64(totalEarlyLeavingMinutes)
		finalSalary := employee.BasicSalary - lateDeduction - earlyLeavingDeduction

		// Memperbarui status pembayaran gaji
		employee.PaidStatus = true
		db.Save(&employee)

		// Membuat catatan pembayaran gaji
		payrollInfo := models.PayrollInfo{
			EmployeeID:       employee.ID,
			BasicSalary:      finalSalary,
			PayslipType:      employee.PaySlipType,
			PaidStatus:       employee.PaidStatus,
			FullNameEmployee: employee.FirstName + " " + employee.LastName,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		db.Create(&payrollInfo)

		// Mereset total menit keterlambatan dan total menit early leaving untuk bulan berikutnya
		db.Model(&models.Attendance{}).Where("employee_id = ?", employee.ID).Updates(map[string]interface{}{"late_minutes": 0, "early_leaving_minutes": 0})

		// Mengirim notifikasi email tentang pembayaran gaji
		go func(email, fullName string, finalSalary float64) {
			if err := helper.SendSalaryTransferNotification(email, fullName, finalSalary); err != nil {
				fmt.Println("Failed to send salary transfer notification email:", err)
			}
		}(employee.Email, employee.FirstName+" "+employee.LastName, finalSalary)

		// Membuat response sukses
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Paid status updated successfully",
			"employee": map[string]interface{}{
				"id":             employee.ID,
				"payroll_id":     employee.PayrollID,
				"first_name":     employee.FirstName,
				"last_name":      employee.LastName,
				"full_name":      employee.FullName,
				"contact_number": employee.ContactNumber,
				"gender":         employee.Gender,
				"email":          employee.Email,
				"username":       employee.Username,
				"shift_id":       employee.ShiftID,
				"shift":          employee.Shift,
				"role_id":        employee.RoleID,
				"role":           employee.Role,
				"department_id":  employee.DepartmentID,
				"department":     employee.Department,
				"designation_id": employee.DesignationID,
				"designation":    employee.Designation,
				"basic_salary":   employee.BasicSalary,
				"hourly_rate":    employee.HourlyRate,
				"pay_slip_type":  employee.PaySlipType,
				"is_active":      employee.IsActive,
				"paid_status":    employee.PaidStatus,
				"created_at":     employee.CreatedAt,
				"updated_at":     employee.UpdatedAt,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

/*
// UpdatePaidStatusByPayrollID merupakan handler untuk memperbarui status pembayaran gaji berdasarkan ID payroll
func UpdatePaidStatusByPayrollID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Mendapatkan token dari header Authorization
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Memeriksa format token
		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		// Memverifikasi token
		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mendapatkan data admin berdasarkan username
		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Memeriksa apakah admin memiliki akses sebagai admin HR
		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		// Mendapatkan ID payroll dari URL parameter
		payrollID := c.Param("payroll_id")

		// Mendapatkan data employee berdasarkan payrollID
		var employee models.Employee
		if err := db.Preload("PayrollInfo").Where("payroll_id = ?", payrollID).First(&employee).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Mendapatkan bulan dan tahun saat ini
		currentMonth := time.Now().Format("2006-01")

		// Mendapatkan semua data kehadiran employee pada bulan dan tahun saat ini
		var attendances []models.Attendance
		result = db.Where("employee_id = ? AND attendance_date LIKE ?", employee.ID, currentMonth+"%").Find(&attendances)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusInternalServerError, Error: true, Message: "Failed to fetch attendances"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Menghitung total menit keterlambatan dari semua kehadiran pada bulan ini
		totalLateMinutes := 0
		for _, attendance := range attendances {
			totalLateMinutes += attendance.LateMinutes
		}

		// Menghitung potongan gaji berdasarkan total menit keterlambatan
		deduction := (employee.HourlyRate / 60) * float64(totalLateMinutes)
		finalSalary := employee.BasicSalary - deduction

		// Memperbarui status pembayaran gaji
		employee.PaidStatus = true
		db.Save(&employee)

		// Membuat catatan pembayaran gaji
		payrollInfo := models.PayrollInfo{
			EmployeeID:       employee.ID,
			BasicSalary:      finalSalary,
			PayslipType:      employee.PaySlipType,
			PaidStatus:       employee.PaidStatus,
			FullNameEmployee: employee.FirstName + " " + employee.LastName,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		db.Create(&payrollInfo)

		// Mereset total menit keterlambatan untuk bulan berikutnya
		db.Model(&models.Attendance{}).Where("employee_id = ?", employee.ID).Update("late_minutes", 0)

		// Mengirim notifikasi email tentang pembayaran gaji
		go func(email, fullName string, finalSalary float64) {
			if err := helper.SendSalaryTransferNotification(email, fullName, finalSalary); err != nil {
				fmt.Println("Failed to send salary transfer notification email:", err)
			}
		}(employee.Email, employee.FirstName+" "+employee.LastName, finalSalary)

		// Membuat response sukses
		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Paid status updated successfully",
			"employee": map[string]interface{}{
				"id":             employee.ID,
				"payroll_id":     employee.PayrollID,
				"first_name":     employee.FirstName,
				"last_name":      employee.LastName,
				"full_name":      employee.FullName,
				"contact_number": employee.ContactNumber,
				"gender":         employee.Gender,
				"email":          employee.Email,
				"username":       employee.Username,
				"shift_id":       employee.ShiftID,
				"shift":          employee.Shift,
				"role_id":        employee.RoleID,
				"role":           employee.Role,
				"department_id":  employee.DepartmentID,
				"department":     employee.Department,
				"designation_id": employee.DesignationID,
				"designation":    employee.Designation,
				"basic_salary":   employee.BasicSalary,
				"hourly_rate":    employee.HourlyRate,
				"pay_slip_type":  employee.PaySlipType,
				"is_active":      employee.IsActive,
				"paid_status":    employee.PaidStatus,
				"created_at":     employee.CreatedAt,
				"updated_at":     employee.UpdatedAt,
			},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}
*/

/*
func UpdatePaidStatusByPayrollID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		payrollID := c.Param("payroll_id")

		var employee models.Employee
		if err := db.Preload("PayrollInfo").Where("payroll_id = ?", payrollID).First(&employee).Error; err != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		employee.PaidStatus = true
		db.Save(&employee)

		payrollInfo := models.PayrollInfo{
			EmployeeID:       employee.ID,
			BasicSalary:      employee.BasicSalary,
			PayslipType:      employee.PaySlipType,
			PaidStatus:       employee.PaidStatus,
			FullNameEmployee: employee.FirstName + " " + employee.LastName,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		db.Create(&payrollInfo)

		go func(email, fullName string, basicSalary float64) {
			if err := helper.SendSalaryTransferNotification(email, fullName, basicSalary); err != nil {
				fmt.Println("Failed to send salary transfer notification email:", err)
			}
		}(employee.Email, employee.FirstName+" "+employee.LastName, employee.BasicSalary)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Paid status updated successfully",
			"employee": map[string]interface{}{
				"id":             employee.ID,
				"payroll_id":     employee.PayrollID,
				"first_name":     employee.FirstName,
				"last_name":      employee.LastName,
				"full_name":      employee.FullName,
				"contact_number": employee.ContactNumber,
				"gender":         employee.Gender,
				"email":          employee.Email,
				"username":       employee.Username,
				"shift_id":       employee.ShiftID,
				"shift":          employee.Shift,
				"role_id":        employee.RoleID,
				"role":           employee.Role,
				"department_id":  employee.DepartmentID,
				"department":     employee.Department,
				"designation_id": employee.DesignationID,
				"designation":    employee.Designation,
				"basic_salary":   employee.BasicSalary,
				"hourly_rate":    employee.HourlyRate,
				"pay_slip_type":  employee.PaySlipType,
				"is_active":      employee.IsActive,
				"paid_status":    employee.PaidStatus,
				"created_at":     employee.CreatedAt,
				"updated_at":     employee.UpdatedAt,
			},
		}
		return c.JSON(http.StatusOK, successResponse)

	}
}
*/

func GetAllPayrollHistory(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"code": http.StatusUnauthorized, "error": true, "message": "Authorization token is missing"})
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"code": http.StatusUnauthorized, "error": true, "message": "Invalid token format"})
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"code": http.StatusUnauthorized, "error": true, "message": "Invalid token"})
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"code": http.StatusNotFound, "error": true, "message": "Admin user not found"})
		}

		if !adminUser.IsAdminHR {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"code": http.StatusForbidden, "error": true, "message": "Access denied"})
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

		searching := c.QueryParam("searching")

		query := db.Model(&models.PayrollInfo{})

		if searching != "" {
			searchPattern := "%" + searching + "%"
			query = query.Where("full_name_employee ILIKE ?", searchPattern)
		}

		var totalCount int64
		query.Count(&totalCount)

		var payrollInfoList []models.PayrollInfo
		if err := query.Order("id DESC").Offset(offset).Limit(perPage).Find(&payrollInfoList).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching payroll information"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":              http.StatusOK,
			"error":             false,
			"message":           "Payroll information retrieved successfully",
			"payroll_info_list": payrollInfoList,
			"pagination": map[string]interface{}{
				"total_count": totalCount,
				"page":        page,
				"per_page":    perPage,
			},
		})
	}
}

func CreateAdvanceSalaryByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		var advanceSalary models.AdvanceSalary
		if err := c.Bind(&advanceSalary); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if advanceSalary.EmployeeID == 0 || advanceSalary.MonthAndYear == "" || advanceSalary.Amount == 0 || advanceSalary.Reason == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid data. Employee ID, Month and Year, Amount, and Reason are required fields"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var employee models.Employee
		result = db.First(&employee, advanceSalary.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		advanceSalary.FullnameEmployee = employee.FullName
		advanceSalary.Emi = advanceSalary.MonthlyInstallmentAmt

		advanceSalary.Status = "Pending"

		if advanceSalary.OneTimeDeduct == "Yes" {
			advanceSalary.MonthlyInstallmentAmt = advanceSalary.Amount
			advanceSalary.Emi = advanceSalary.MonthlyInstallmentAmt
		}

		advanceSalary.Emi = advanceSalary.MonthlyInstallmentAmt

		_, err = time.Parse("2006-01", advanceSalary.MonthAndYear)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid date format. Required format: yyyy-mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		db.Create(&advanceSalary)

		// Mengirim notifikasi email kepada karyawan terkait
		err = helper.SendAdvanceSalaryNotification(employee.Email, advanceSalary.FullnameEmployee, advanceSalary.MonthAndYear, advanceSalary.Amount, advanceSalary.OneTimeDeduct, advanceSalary.MonthlyInstallmentAmt, advanceSalary.Reason)
		if err != nil {
			fmt.Println("Gagal mengirim email notifikasi advance salary:", err)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Advance Salary created successfully",
			"data":    advanceSalary,
		}
		return c.JSON(http.StatusCreated, successResponse)

	}
}

func GetAllAdvanceSalariesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
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

		searching := c.QueryParam("searching")

		query := db.Model(&models.AdvanceSalary{})

		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where("LOWER(fullname_employee) LIKE ? OR amount = ? OR LOWER(status) LIKE ?", searchPattern, helper.ParseStringToInt(searching), searchPattern)
		}

		var totalCount int64
		query.Count(&totalCount)

		var advanceSalaries []models.AdvanceSalary
		if err := query.Order("id DESC").Offset(offset).Limit(perPage).Find(&advanceSalaries).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching advance salaries"})
		}

		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Advance Salary history retrieved successfully",
			"data":       advanceSalaries,
			"pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetAdvanceSalaryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		id := c.Param("id")

		advanceSalaryID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var advanceSalary models.AdvanceSalary
		result = db.First(&advanceSalary, advanceSalaryID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Advance Salary not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Advance Salary retrieved successfully",
			"data":    advanceSalary,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateAdvanceSalaryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		id := c.Param("id")

		advanceSalaryID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var advanceSalary models.AdvanceSalary
		result = db.First(&advanceSalary, advanceSalaryID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Advance Salary not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedData models.AdvanceSalary
		if err := c.Bind(&updatedData); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedData.Amount != 0 {
			advanceSalary.Amount = updatedData.Amount
			advanceSalary.Emi = updatedData.Amount
			advanceSalary.MonthlyInstallmentAmt = updatedData.Amount
		}

		if updatedData.OneTimeDeduct != "" {
			advanceSalary.OneTimeDeduct = updatedData.OneTimeDeduct
		}
		if updatedData.MonthlyInstallmentAmt != 0 {
			advanceSalary.MonthlyInstallmentAmt = updatedData.MonthlyInstallmentAmt
		}
		if updatedData.Reason != "" {
			advanceSalary.Reason = updatedData.Reason
		}
		if updatedData.Emi != 0 {
			advanceSalary.Emi = updatedData.Emi
		}
		if updatedData.Paid != 0 {
			advanceSalary.Paid = updatedData.Paid
		}

		if updatedData.Status != "" {
			advanceSalary.Status = updatedData.Status
		}

		if updatedData.EmployeeID != 0 {
			var employee models.Employee
			result := db.First(&employee, updatedData.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			advanceSalary.EmployeeID = updatedData.EmployeeID
			advanceSalary.FullnameEmployee = employee.FullName
		}

		db.Save(&advanceSalary)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Advance Salary updated successfully",
			"data":    advanceSalary,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteAdvanceSalaryByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		id := c.Param("id")

		advanceSalaryID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var advanceSalary models.AdvanceSalary
		result = db.First(&advanceSalary, advanceSalaryID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Advance Salary not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&advanceSalary)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Advance Salary deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func CreateRequestLoanByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		var requestLoan models.RequestLoan
		if err := c.Bind(&requestLoan); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if requestLoan.EmployeeID == 0 || requestLoan.MonthAndYear == "" || requestLoan.Amount == 0 || requestLoan.Reason == "" {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid data. Employee ID, Month and Year, Amount, and Reason are required fields"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var employee models.Employee
		result = db.First(&employee, requestLoan.EmployeeID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		requestLoan.FullnameEmployee = employee.FullName
		requestLoan.Emi = requestLoan.MonthlyInstallmentAmt

		requestLoan.Status = "Pending"

		requestLoan.Remaining = requestLoan.Amount - requestLoan.Paid

		if requestLoan.OneTimeDeduct == "Yes" {
			requestLoan.MonthlyInstallmentAmt = requestLoan.Amount
		}

		_, err = time.Parse("2006-01", requestLoan.MonthAndYear)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid date format. Required format: yyyy-mm"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		db.Create(&requestLoan)

		// Mengirim notifikasi email kepada karyawan
		err = helper.SendRequestLoanNotification(employee.Email, employee.FullName, requestLoan.MonthAndYear, requestLoan.Amount, requestLoan.OneTimeDeduct, requestLoan.MonthlyInstallmentAmt, requestLoan.Reason)
		if err != nil {
			fmt.Println("Failed to send request loan notification email:", err)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusCreated,
			"error":   false,
			"message": "Request Loan created successfully",
			"data":    requestLoan,
		}
		return c.JSON(http.StatusCreated, successResponse)

	}
}

func GetAllRequestLoanByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
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

		searching := c.QueryParam("searching")

		query := db.Model(&models.RequestLoan{})

		if searching != "" {
			searchPattern := "%" + strings.ToLower(searching) + "%"
			query = query.Where("LOWER(fullname_employee) LIKE ? OR amount = ? OR LOWER(status) LIKE ?", searchPattern, helper.ParseStringToInt(searching), searchPattern)
		}

		var totalCount int64
		query.Count(&totalCount)

		var requestLoans []models.RequestLoan
		if err := query.Order("id DESC").Offset(offset).Limit(perPage).Find(&requestLoans).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "message": "Error fetching request loans"})
		}

		successResponse := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Request Loan history retrieved successfully",
			"data":       requestLoans,
			"pagination": map[string]interface{}{"total_count": totalCount, "page": page, "per_page": perPage},
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func GetRequestLoanByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		id := c.Param("id")

		requestLoanID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var requestLoan models.RequestLoan
		result = db.First(&requestLoan, requestLoanID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Request Loan not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Request Loan retrieved successfully",
			"data":    requestLoan,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func UpdateRequestLoanByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		id := c.Param("id")

		requestLoanID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var requestLoan models.RequestLoan
		result = db.First(&requestLoan, requestLoanID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Request Loan not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedData models.RequestLoan
		if err := c.Bind(&updatedData); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedData.Amount != 0 {
			requestLoan.Amount = updatedData.Amount
			requestLoan.Emi = updatedData.Amount
			requestLoan.MonthlyInstallmentAmt = updatedData.Amount
			requestLoan.Remaining = updatedData.Amount - requestLoan.Paid
		}

		if updatedData.OneTimeDeduct != "" {
			requestLoan.OneTimeDeduct = updatedData.OneTimeDeduct
		}
		if updatedData.MonthlyInstallmentAmt != 0 {
			requestLoan.MonthlyInstallmentAmt = updatedData.MonthlyInstallmentAmt
		}
		if updatedData.Reason != "" {
			requestLoan.Reason = updatedData.Reason
		}
		if updatedData.Emi != 0 {
			requestLoan.Emi = updatedData.Emi
		}
		if updatedData.Paid != 0 {
			requestLoan.Paid = updatedData.Paid
			requestLoan.Remaining = requestLoan.Amount - updatedData.Paid
		}

		if updatedData.Status != "" {
			requestLoan.Status = updatedData.Status
		}

		if updatedData.EmployeeID != 0 {
			var employee models.Employee
			result := db.First(&employee, updatedData.EmployeeID)
			if result.Error != nil {
				errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Employee not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			requestLoan.EmployeeID = updatedData.EmployeeID
			requestLoan.FullnameEmployee = employee.FullName
		}

		db.Save(&requestLoan)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Request Loan updated successfully",
			"data":    requestLoan,
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteRequestLoanByIDByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		username, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusUnauthorized, Error: true, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var adminUser models.Admin
		result := db.Where("username = ?", username).First(&adminUser)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Admin user not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if !adminUser.IsAdminHR {
			errorResponse := helper.Response{Code: http.StatusForbidden, Error: true, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		id := c.Param("id")

		requestLoanID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid ID format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var requestLoan models.RequestLoan
		result = db.First(&requestLoan, requestLoanID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Request Loan not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&requestLoan)

		successResponse := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Request Loan deleted successfully",
		}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func ResetPaidStatus(db *gorm.DB) {
	// Get the current date
	currentDate := time.Now().Format("2006-01-02")
	fmt.Printf("Running ResetPaidStatus on %s\n", currentDate)

	// Update the paid_status of all employees to false
	result := db.Model(&models.Employee{}).Where("paid_status = ?", true).Update("paid_status", false)
	if result.Error != nil {
		fmt.Printf("Failed to reset paid status: %v\n", result.Error)
		return
	}

	fmt.Println("Successfully reset paid status for all employees.")
}
