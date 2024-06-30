package models

import "time"

type Announcement struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Title          string     `json:"title"`
	DepartmentID   uint       `json:"department_id"`
	DepartmentName string     `json:"department_name"`
	Summary        string     `json:"summary"`
	Description    string     `json:"description"`
	StartDate      string     `json:"start_date"`
	EndDate        string     `json:"end_date"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      time.Time
}
