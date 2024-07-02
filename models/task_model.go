package models

import (
	"time"
)

type Task struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	Title         string `json:"title"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	EstimatedHour int    `json:"estimated_hour"`
	ProjectID     uint   `json:"project_id"`
	//Project       Project    `gorm:"foreignKey:ProjectID;references:ID" json:"project"`
	ProjectName string     `json:"project_name"`
	Summary     string     `json:"summary"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	ProgressBar int        `json:"progress_bar"`
	Notes       []Note     `json:"notes" gorm:"foreignKey:TaskID"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Note struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	TaskID    uint       `json:"task_id"`
	NoteText  string     `json:"note_text"`
	Fullname  string     `json:"fullname"`
	CreatedAt *time.Time `json:"created_at"`
}
