package models

import "time"

type Attendance struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	EmployeeID     uint       `json:"employee_id"`
	Username       string     `json:"username"`
	AttendanceDate string     `json:"attendance_date"` // Format: yyyy-mm-dd
	InTime         string     `json:"in_time"`
	OutTime        string     `json:"out_time"`
	CreatedAt      *time.Time `json:"created_at"`
}
