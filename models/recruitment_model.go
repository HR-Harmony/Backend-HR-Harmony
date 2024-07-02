package models

import "time"

type NewJob struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	Title         string `json:"title"`
	JobType       string `json:"job_type"`
	DesignationID uint   `json:"designation_id"`
	//Designation       Designation `gorm:"foreignKey:DesignationID;references:ID" json:"designation"`
	DesignationName   string    `json:"designation_name"`
	NumberOfPosition  int       `json:"number_of_position"`
	IsPublish         bool      `json:"is_publish"`
	DateClosing       string    `json:"date_closing"`
	MinimumExperience string    `json:"minimum_experience"`
	ShortDescription  string    `json:"short_description"`
	LongDescription   string    `json:"long_description"`
	CreatedAt         time.Time `json:"created_at"`
}
