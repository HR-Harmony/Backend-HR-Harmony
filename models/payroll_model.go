// models/payroll_info.go

package models

import "time"

type PayrollInfo struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	EmployeeID       uint      `json:"employee_id"`
	FullNameEmployee string    `json:"full_name_employee"`
	BasicSalary      float64   `json:"basic_salary"`
	PayslipType      string    `json:"payslip_type"`
	PaidStatus       bool      `json:"paid_status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AdvanceSalary struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	EmployeeID            uint      `json:"employee_id"`
	FullnameEmployee      string    `json:"fullname_employee"`
	MonthAndYear          string    `json:"month_and_year"` // Format: yyyy-mm
	Amount                int       `json:"amount"`
	OneTimeDeduct         string    `json:"one_time_deduct"`
	MonthlyInstallmentAmt int       `json:"monthly_installment_amount"`
	Reason                string    `json:"reason"`
	Emi                   int       `json:"emi"`
	Paid                  int       `json:"paid"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
}

type RequestLoan struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	EmployeeID            uint      `json:"employee_id"`
	FullnameEmployee      string    `json:"fullname_employee"`
	MonthAndYear          string    `json:"month_and_year"` // Format: yyyy-mm
	Amount                int       `json:"amount"`
	OneTimeDeduct         string    `json:"one_time_deduct"`
	MonthlyInstallmentAmt int       `json:"monthly_installment_amount"`
	Reason                string    `json:"reason"`
	Emi                   int       `json:"emi"`
	Paid                  int       `json:"paid"`
	Status                string    `json:"status"`
	Remaining             int       `json:"remaining"`
	CreatedAt             time.Time `json:"created_at"`
}
