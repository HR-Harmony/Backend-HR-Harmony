package models

import "time"

type GoalType struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	GoalType  string    ` json:"goal_type"`
	CreatedAt time.Time `json:"created_at"`
	//Goal      []Goal    `gorm:"foreignKey:GoalTypeID;references:ID" json:"goal"`
}

type Goal struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	GoalTypeID uint `json:"goal_type_id"`
	//GoalType          GoalType      `gorm:"foreignKey:GoalTypeID;references:ID" json:"goal_type"`
	GoalTypeName string `json:"goal_type_name"`
	ProjectID    uint   `json:"project_id"`
	//Project           Project       `gorm:"foreignKey:ProjectID;references:ID" json:"project"`
	ProjectName string `json:"project_name"`
	TaskID      uint   `json:"task_id"`
	Task        Task   `gorm:"foreignKey:TaskID;references:ID" json:"task"`
	TaskName    string `json:"task_name"`
	TrainingID  uint   `json:"training_id"`
	//Training          Training      `gorm:"foreignKey:TrainingID ;references:ID" json:"training"`
	TrainingSkillID uint `json:"training_skill_id"`
	//TrainingSkill     TrainingSkill `gorm:"foreignKey:TrainingSkillID ;references:ID" json:"training_skill"`
	TrainingSkillName string `json:"training_skill_name"`
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
