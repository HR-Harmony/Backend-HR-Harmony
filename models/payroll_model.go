// models/payroll_info.go

package models

import "time"

type PayrollInfo struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	EmployeeID  uint      `json:"employee_id"`
	BasicSalary float64   `json:"basic_salary"`
	PayslipType string    `json:"payslip_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
