package models

import "time"

type Exit struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	ExitName  string     `json:"exit_name"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt time.Time
}

type ExitEmployee struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	EmployeeID     uint      `json:"employee_id"`
	ExitID         uint      `json:"exit_id"`
	DisableAccount bool      `json:"disable_account"`
	Description    string    `json:"description"`
	ExitDate       string    `json:"exit_date"`
	CreatedAt      time.Time `json:"created_at"`
}
