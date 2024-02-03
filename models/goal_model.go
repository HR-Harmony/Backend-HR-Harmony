package models

import "time"

type GoalType struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	GoalType  string    ` json:"goal_type"`
	CreatedAt time.Time `json:"created_at"`
}
