package models

import "time"

type LeaveRequestType struct {
	ID                 uint   `gorm:"primaryKey" json:"id"`
	LeaveType          string `json:"leave_type"`
	DaysPerYears       int    `json:"days_per_years"`
	IsRequiresApproval bool   `json:"is_requires_approval"`
}

type LeaveRequest struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	EmployeeID  uint       `json:"employee_id"`
	Username    string     `json:"username"`
	LeaveTypeID uint       `json:"leave_type_id"`
	LeaveType   string     `json:"leave_type"`
	StartDate   string     `json:"start_date"`
	EndDate     string     `json:"end_date"`
	IsHalfDay   bool       `json:"is_half_day"`
	Remarks     string     `json:"remarks"`
	LeaveReason string     `json:"leave_reason"`
	Days        int        `json:"days"`
	Status      string     `json:"status"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
