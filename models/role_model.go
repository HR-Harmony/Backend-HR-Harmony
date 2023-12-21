package models

import "time"

type Role struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	RoleName  string     `json:"role_name"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt time.Time
}
