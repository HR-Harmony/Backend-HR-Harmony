package models

type LeaveRequestType struct {
	ID                 uint   `gorm:"primaryKey" json:"id"`
	LeaveType          string `json:"leave_type"`
	DaysPerYears       int    `json:"days_per_years"`
	IsRequiresApproval bool   `json:"is_requires_approval"`
}
