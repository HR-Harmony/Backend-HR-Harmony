package models

import "time"

type Helpdesk struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Subject      string `json:"subject"`
	Priority     string `json:"priority"`
	DepartmentID uint   `json:"department_id"`
	//Department       Department `gorm:"foreignKey:DepartmentID;references:ID" json:"department"`
	DepartmentName string `json:"department_name"`
	EmployeeID     uint   `json:"employee_id"`
	//Employee         Employee   `gorm:"foreignKey:EmployeeID;references:ID" json:"employee"`
	EmployeeUsername string     `json:"employee_username"`
	EmployeeFullName string     `json:"employee_full_name"`
	Description      string     `json:"description"`
	Status           string     `json:"status"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
