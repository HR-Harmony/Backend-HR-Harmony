package models

import "time"

type ResetPasswordOTP struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	EmployeeID  uint      `json:"employee_id"`
	OTP         string    `json:"otp"`
	IsUsed      bool      `json:"is_used" gorm:"default:false"`
	ExpiredAt   time.Time `json:"expired_at"`
	RequestedAt time.Time `json:"requested_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type AdminResetPasswordOTP struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AdminID     uint      `json:"admin_id"`
	OTP         string    `json:"otp"`
	ExpiredAt   time.Time `json:"expired_at"`
	IsUsed      bool      `gorm:"default:false" json:"is_used"`
	RequestedAt time.Time `json:"requested_at"`
	CreatedAt   time.Time `json:"created_at"`
}
