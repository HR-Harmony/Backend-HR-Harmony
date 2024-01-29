package models

import "time"

type Helpdesk struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	Subject          string     `json:"subject"`
	Priority         string     `json:"priority"`
	DepartmentID     uint       `json:"department_id"`
	DepartmentName   string     `json:"department_name"`
	EmployeeID       uint       `json:"employee_id"`
	EmployeeUsername string     `json:"employee_username"`
	Description      string     `json:"description"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
