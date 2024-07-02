package models

import "time"

type Department struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	DepartmentName string     `json:"department_name"`
	EmployeeID     uint       `json:"employee_id"` // tambahan
	FullName       string     `json:"full_name"`   // tambahan
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      time.Time
	Designations   []Designation `gorm:"foreignKey:DepartmentID" json:"designations"` // relasi balik
	//Employee       []Employee    `gorm:"foreignKey:DepartmentID;references:ID" json:"employee"`
}
