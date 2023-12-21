package models

import "time"

type Employee struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	ContactNumber string     `json:"contact_number"`
	Gender        string     `json:"gender"`
	Email         string     `json:"email"`
	Username      string     `json:"username"`
	Password      string     `json:"password"`
	ShiftID       uint       `json:"shift_id"`
	Shift         string     `json:"shift"`
	RoleID        uint       `json:"role_id"`
	Role          string     `json:"role"`
	DepartmentID  uint       `json:"department_id"`
	Department    string     `json:"department"`
	BasicSalary   float64    `json:"basic_salary"`
	HourlyRate    float64    `json:"hourly_rate"`
	PaySlipType   string     `json:"pay_slip_type"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     time.Time
}
