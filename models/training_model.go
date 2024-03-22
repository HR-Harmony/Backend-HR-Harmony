package models

import "time"

type Trainer struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	FullName      string     `json:"full_name"`
	ContactNumber string     `json:"contact_number"`
	Email         string     `json:"email"`
	Expertise     string     `json:"expertise"`
	Address       string     `json:"address"`
	CreatedAt     *time.Time `json:"created_at"`
}

type TrainingSkill struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	TrainingSkill string `json:"training_skill"`
	CreatedAt     *time.Time
	UpdateAt      time.Time
}
