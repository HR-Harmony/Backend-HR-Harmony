package models

import "time"

type Attendance struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	EmployeeID          uint       `json:"employee_id"`
	Username            string     `json:"username"`
	FullNameEmployee    string     `json:"full_name_employee"`
	AttendanceDate      string     `json:"attendance_date"` // Format: yyyy-mm-dd
	InTime              string     `json:"in_time"`
	OutTime             string     `json:"out_time"`
	TotalWork           string     `json:"total_work"`
	Status              string     `json:"status"`
	Late                string     `json:"late"`         // New field for late time
	LateMinutes         int        `json:"late_minutes"` // New field for late time in minutes
	EarlyLeaving        string     `json:"early_leaving"`
	EarlyLeavingMinutes int        `json:"early_leaving_minutes"` // New field for early leaving time in minutes
	CreatedAt           *time.Time `json:"created_at"`
}

type OvertimeRequest struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	EmployeeID       uint       `json:"employee_id"`
	Username         string     `json:"username"`
	FullNameEmployee string     `json:"full_name_employee"`
	Date             string     `json:"date"` // Format: yyyy-mm-dd
	InTime           string     `json:"in_time"`
	OutTime          string     `json:"out_time"`
	Reason           string     `json:"reason"`
	TotalWork        string     `json:"total_work"`
	TotalMinutes     int        `json:"total_minutes"` // New field for total minutes
	Status           string     `json:"status"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
