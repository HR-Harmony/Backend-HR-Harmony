package models

import "time"

type Department struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	DepartmentName string     `json:"department_name"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      time.Time
}
