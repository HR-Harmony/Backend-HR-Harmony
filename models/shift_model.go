package models

import "time"

type Shift struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	ShiftName        string     `json:"shift_name"`
	MondayInTime     string     `json:"monday_in_time"`
	MondayOutTime    string     `json:"monday_out_time"`
	TuesdayInTime    string     `json:"tuesday_in_time"`
	TuesdayOutTime   string     `json:"tuesday_out_time"`
	WednesdayInTime  string     `json:"wednesday_in_time"`
	WednesdayOutTime string     `json:"wednesday_out_time"`
	ThursdayInTime   string     `json:"thursday_in_time"`
	ThursdayOutTime  string     `json:"thursday_out_time"`
	FridayInTime     string     `json:"friday_in_time"`
	FridayOutTime    string     `json:"friday_out_time"`
	SaturdayInTime   string     `json:"saturday_in_time"`
	SaturdayOutTime  string     `json:"saturday_out_time"`
	SundayInTime     string     `json:"sunday_in_time"`
	SundayOutTime    string     `json:"sunday_out_time"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdatedAt        time.Time
}
