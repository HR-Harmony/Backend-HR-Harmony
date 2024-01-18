package models

import "time"

type Policy struct {
	ID                     uint       `gorm:"primaryKey" json:"id"`
	Title                  string     `json:"title"`
	Description            string     `json:"description"`
	CreatedByAdminID       uint       `json:"created_by_admin_id"`
	CreatedByAdminUsername string     `json:"created_by_admin_username"`
	CreatedAt              *time.Time `json:"created_at"`
}
