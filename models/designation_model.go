package models

import "time"

type Designation struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	DepartmentID    uint       `json:"department_id"`
	Department      Department `gorm:"references:ID" json:"department"`
	DepartmentName  string     `json:"department_name"`
	DesignationName string     `json:"designation_name"`
	Description     string     `json:"description"`
	CreatedAt       time.Time  `json:"created_at"`
	Employee        []Employee `gorm:"foreignKey:DesignationID;references:ID" json:"employee"`
}
