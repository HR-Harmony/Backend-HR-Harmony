package models

import "time"

type Exit struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	ExitName     string     `json:"exit_name"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    time.Time
	ExitEmployee []ExitEmployee `gorm:"foreignKey:ExitID;references:ID" json:"exit_employee"`
}

type ExitEmployee struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	EmployeeID       uint      `json:"employee_id"`
	FullNameEmployee string    `json:"full_name_employee"`
	ExitID           uint      `json:"exit_id"`
	Exit             Exit      `gorm:"foreignKey:ExitID;references:ID" json:"exit"`
	ExitName         string    `json:"exit_name"`
	DisableAccount   bool      `json:"disable_account"`
	ExitInterview    string    `json:"exit_interview"`
	Description      string    `json:"description"`
	ExitDate         string    `json:"exit_date"`
	CreatedAt        time.Time `json:"created_at"`
}
