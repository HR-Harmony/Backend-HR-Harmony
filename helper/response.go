package helper

import (
	"hrsale/models"
	"time"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ResponseShift struct {
	Code    int           `json:"code"`
	Error   bool          `json:"error"`
	Message string        `json:"message"`
	Shift   *models.Shift `json:"shift,omitempty"`
}

type EmployeeResponse struct {
	ID            uint       `json:"id"`
	PayrollID     int64      `json:"payroll_id"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	ContactNumber string     `json:"contact_number"`
	Gender        string     `json:"gender"`
	Email         string     `json:"email"`
	Username      string     `json:"username"`
	Password      string     `json:"password"`
	ShiftID       uint       `json:"shift_id"`
	Shift         string     `json:"shift"`
	RoleID        uint       `json:"role_id"`
	Role          string     `json:"role"`
	DepartmentID  uint       `json:"department_id"`
	Department    string     `json:"department"`
	DesignationID uint       `json:"designation_id"`
	Designation   string     `json:"designation"`
	BasicSalary   float64    `json:"basic_salary"`
	HourlyRate    float64    `json:"hourly_rate"`
	PaySlipType   string     `json:"pay_slip_type"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	PaidStatus    bool       `json:"paid_status" gorm:"default:false"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     time.Time  `json:"UpdatedAt"`
}

type Response struct {
	Code                 int                      `json:"code"`
	Error                bool                     `json:"error"`
	Message              string                   `json:"message"`
	Shift                *models.Shift            `json:"shift,omitempty"`
	Shifts               []models.Shift           `json:"shifts,omitempty"`
	Role                 *models.Role             `json:"role,omitempty"`
	Roles                []models.Role            `json:"roles,omitempty"`
	Department           *models.Department       `json:"department,omitempty"`
	Departments          []models.Department      `json:"departments,omitempty"`
	Employee             *models.Employee         `json:"employee,omitempty"`
	Employees            []models.Employee        `json:"employees,omitempty"`
	Exit                 *models.Exit             `json:"exit,omitempty"`
	Exits                []models.Exit            `json:"exits,omitempty"`
	ExitEmployee         *models.ExitEmployee     `json:"exit_employee,omitempty"`
	ExitEmployees        []models.ExitEmployee    `json:"exit_employees,omitempty"`
	Designation          *models.Designation      `json:"designation,omitempty"`
	Designations         []models.Designation     `json:"designations,omitempty"`
	Policy               *models.Policy           `json:"policy,omitempty"`
	Policies             []models.Policy          `json:"policies,omitempty"`
	Admin                *models.Admin            `json:"admin,omitempty"`
	Admins               []models.Admin           `json:"admins,omitempty"`
	Announcement         *models.Announcement     `json:"announcement,omitempty"`
	Announcements        []models.Announcement    `json:"announcements,omitempty"`
	Project              *models.Project          `json:"project,omitempty"`
	Projects             []models.Project         `json:"projects,omitempty"`
	Task                 *models.Task             `json:"task,omitempty"`
	Tasks                []models.Task            `json:"tasks,omitempty"`
	Case                 *models.Case             `json:"case,omitempty"`
	Cases                []models.Case            `json:"cases,omitempty"`
	Disciplinary         *models.Disciplinary     `json:"disciplinary,omitempty"`
	Disciplinaries       []models.Disciplinary    `json:"disciplinaries,omitempty"`
	Helpdesk             *models.Helpdesk         `json:"helpdesk,omitempty"`
	Helpdesks            []models.Helpdesk        `json:"helpdesks,omitempty"`
	PayrollInfo          []map[string]interface{} `json:"payroll_info,omitempty"`
	PayrollID            uint                     `json:"payroll_id,omitempty"`
	PayrollInfoHistorie  *models.PayrollInfo      `json:"payroll_info_historie,omitempty"`
	PayrollInfoHistories []models.PayrollInfo     `json:"payroll_info_histories,omitempty"`
}
