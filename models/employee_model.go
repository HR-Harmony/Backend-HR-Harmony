package models

import "time"

type Employee struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	PayrollID     int64       `json:"payroll_id"`
	FirstName     string      `json:"first_name"`
	LastName      string      `json:"last_name"`
	FullName      string      `json:"full_name"`
	ContactNumber string      `json:"contact_number"`
	Gender        string      `json:"gender"`
	Email         string      `json:"email"`
	Username      string      `json:"username"`
	Password      string      `json:"password"`
	ShiftID       uint        `json:"shift_id"`
	Shift         string      `json:"shift"`
	RoleID        uint        `json:"role_id"`
	Role          string      `json:"role"`
	DepartmentID  uint        `json:"department_id"`
	Department    string      `json:"department"`
	DesignationID uint        `json:"designation_id"`
	Designation   string      `json:"designation"`
	BasicSalary   float64     `json:"basic_salary"`
	HourlyRate    float64     `json:"hourly_rate"`
	PaySlipType   string      `json:"pay_slip_type"`
	IsClient      bool        `json:"is_client" gorm:"default:false"`
	IsActive      bool        `json:"is_active" gorm:"default:true"`
	IsExit        bool        `json:"is_exit" gorm:"default:false"`
	PaidStatus    bool        `json:"paid_status" gorm:"default:false"`
	Country       string      `json:"country"`
	PayrollInfo   PayrollInfo `gorm:"foreignKey:EmployeeID"`

	// Details Employee - Basic Information
	MaritalStatus string `json:"marital_status"`
	Religion      string `json:"religion"`
	BloodGroup    string `json:"blood_group"`
	Nationality   string `json:"nationality"`
	Citizenship   string `json:"citizenship"`
	BpjsKesehatan string `json:"bpjs_kesehatan"`
	Address1      string `json:"address1"`
	Address2      string `json:"address2"`
	City          string `json:"city"`
	StateProvince string `json:"state_province"`
	ZipPostalCode string `json:"zip_postal_code"`
	// Details Employee - Personal Information

	//Bio Employee
	Bio string `json:"bio"`

	// Social Profile Employee
	FacebookURL  string `json:"facebook_url"`
	InstagramURL string `json:"instagram_url"`
	TwitterURL   string `json:"twitter_url"`
	LinkedinURL  string `json:"linkedin_url"`

	// Bank Account Employee
	AccountTitle  string `json:"account_title"`
	AccountNumber string `json:"account_number"`
	BankName      string `json:"bank_name"`
	Iban          string `json:"iban"`
	SwiftCode     string `json:"swift_code"`
	BankBranch    string `json:"bank_branch"`

	// Emergency Contact Employee
	EmergencyContactFullName string `json:"emergency_contact_full_name"`
	EmergencyContactNumber   string `json:"emergency_contact_number"`
	EmergencyContactEmail    string `json:"emergency_contact_email"`
	EmergencyContactAddress  string `json:"emergency_contact_address"`

	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt time.Time
}
