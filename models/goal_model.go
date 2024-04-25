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
	ProjectID         uint   `json:"project_id"`
	ProjectName       string `json:"project_name"`
	TaskID            uint   `json:"task_id"`
	TaskName          string `json:"task_name"`
	TrainingID        uint   `json:"training_id"`
	Subject           string `json:"subject"`
	TargetAchievement string `json:"target_achievement"`
	StartDate         string `json:"start_date"`
	EndDate           string `json:"end_date"`
	Description       string `json:"description"`
	GoalRating        uint   `json:"goal_rating"`
	ProgressBar       uint   `json:"progress_bar"`
	Status            string `json:"status"`
	CreatedAt         string `json:"created_at"`
}
