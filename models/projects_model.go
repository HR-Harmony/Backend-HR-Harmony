package models

import (
	"time"
)

type Project struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Title          string     `json:"title"`
	EmployeeID     uint       `json:"employee_id"`
	Username       string     `json:"username"`
	EstimatedHour  int        `json:"estimated_hour"`
	Priority       string     `json:"priority" `
	StartDate      string     `json:"start_date"`
	EndDate        string     `json:"end_date"`
	Summary        string     `json:"summary"`
	DepartmentID   uint       `json:"department_id"`
	DepartmentName string     `json:"department_name"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	ProjectBar     int        `json:"project_bar"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
