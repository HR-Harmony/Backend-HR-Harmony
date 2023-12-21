package models

import "time"

type Shift struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	ShiftName string     `json:"shift_name"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt time.Time
}
