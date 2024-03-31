package models

import (
	"time"
)

type Task struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Title         string     `json:"title"`
	StartDate     string     `json:"start_date"`
	EndDate       string     `json:"end_date"`
	EstimatedHour int        `json:"estimated_hour"`
	ProjectID     uint       `json:"project_id"`
	ProjectName   string     `json:"project_name"`
	Summary       string     `json:"summary"`
	Description   string     `json:"description"`
	Status        string     `json:"status"`
	ProgressBar   int        `json:"progress_bar"`
	Notes         string     `json:"notes"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
