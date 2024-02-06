package models

import "time"

type GoalType struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	GoalType  string    ` json:"goal_type"`
	CreatedAt time.Time `json:"created_at"`
}

type Goal struct {
	ID                uint   `gorm:"primaryKey" json:"id"`
	GoalTypeID        uint   `json:"goal_type_id"`
	GoalTypeName      string `json:"goal_type_name"`
	Subject           string `json:"subject"`
	TargetAchievement string `json:"target_achievement"`
	StartDate         string `json:"start_date"`
	EndDate           string `json:"end_date"`
	Description       string `json:"description"`
	CreatedAt         string `json:"created_at"`
}
