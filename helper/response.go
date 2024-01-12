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
	Code        int                 `json:"code"`
	Error       bool                `json:"error"`
	Message     string              `json:"message"`
	Shift       *models.Shift       `json:"shift,omitempty"`
	Shifts      []models.Shift      `json:"shifts,omitempty"`
	Role        *models.Role        `json:"role,omitempty"`
	Roles       []models.Role       `json:"roles,omitempty"`
	Department  *models.Department  `json:"department,omitempty"`
	Departments []models.Department `json:"departments,omitempty"`
	Employee    *models.Employee    `json:"employee,omitempty"`
}
