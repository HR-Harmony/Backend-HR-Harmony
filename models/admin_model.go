package models

type Admin struct {
	ID                uint    `gorm:"primaryKey" json:"id"`
	FirstName         string  `json:"first_name"`
	LastName          string  `json:"last_name"`
	Fullname          string  `json:"fullname"`
	ContactNumber     string  `json:"contact_number"`
	Gender            string  `json:"gender"`
	Email             string  `json:"email"`
	Username          string  `json:"username"`
	Password          string  `json:"password"`
	Department        string  `json:"department"`
	BasicSalary       float64 `json:"basic_salary"`
	HourlyRate        float64 `json:"hourly_rate"`
	PaySlipType       string  `json:"pay_slip_type"`
	IsAdminHR         bool    `json:"is_admin_hr"`
	IsVerified        bool    `gorm:"default:false" json:"is_verified"`
	VerificationToken string  `json:"verification_token"`
}
