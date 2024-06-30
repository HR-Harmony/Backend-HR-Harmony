package models

import "time"

type Case struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	CaseName  string     `json:"case_name"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt time.Time
}

type Disciplinary struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	EmployeeID       uint       `json:"employee_id"`
	Employee         Employee   `gorm:"foreignKey:EmployeeID;references:ID" json:"employee"`
	UsernameEmployee string     `json:"username_employee"`
	FullNameEmployee string     `json:"full_name_employee"`
	CaseID           uint       `json:"case_id"`
	Case             Case       `gorm:"foreignKey:CaseID;references:ID" json:"case"`
	CaseName         string     `json:"case_name"`
	Subject          string     `json:"subject"`
	CaseDate         string     `json:"case_date"`
	Description      string     `json:"description"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
