package helper

import "hrsale/models"

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

type Response struct {
	Code          int                   `json:"code"`
	Error         bool                  `json:"error"`
	Message       string                `json:"message"`
	Shift         *models.Shift         `json:"shift,omitempty"`
	Shifts        []models.Shift        `json:"shifts,omitempty"`
	Role          *models.Role          `json:"role,omitempty"`
	Roles         []models.Role         `json:"roles,omitempty"`
	Department    *models.Department    `json:"department,omitempty"`
	Departments   []models.Department   `json:"departments,omitempty"`
	Employee      *models.Employee      `json:"employee,omitempty"`
	Exit          *models.Exit          `json:"exit,omitempty"`
	Exits         []models.Exit         `json:"exits,omitempty"`
	ExitEmployee  *models.ExitEmployee  `json:"exit_employee,omitempty"`
	ExitEmployees []models.ExitEmployee `json:"exit_employees,omitempty"`
	Designation   *models.Designation   `json:"designation,omitempty"`
	Designations  []models.Designation  `json:"designations,omitempty"`
	Policy        *models.Policy        `json:"policy,omitempty"`   // Add this line
	Policies      []models.Policy       `json:"policies,omitempty"` // Add this line
	Admin         *models.Admin         `json:"admin,omitempty"`    // Add this line
	Admins        []models.Admin        `json:"admins,omitempty"`   // Add this line
	Announcement  *models.Announcement  `json:"announcement,omitempty"`
	Announcements []models.Announcement `json:"announcements,omitempty"`
}
