package models

import "time"

type Case struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	CaseName  string     `json:"case_name"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt time.Time
}
