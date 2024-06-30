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

type Training struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	TrainerID        uint       `json:"trainer_id"`
	Trainer          Trainer    `gorm:"foreignKey:TrainerID;references:ID" json:"trainer"`
	FullNameTrainer  string     `json:"full_name_trainer"`
	TrainingSkillID  uint       `json:"training_skill_id"`
	TrainingSkill    string     `json:"training_skill"`
	TrainingCost     int        `json:"training_cost"`
	EmployeeID       uint       `json:"employee_id"`
	FullNameEmployee string     `json:"full_name_employee"`
	GoalTypeID       uint       `json:"goal_type_id"`
	GoalType         string     `json:"goal_type"`
	Performance      string     `json:"performance"`
	StartDate        string     `json:"start_date"`
	EndDate          string     `json:"end_date"`
	Status           string     `json:"status"`
	Description      string     `json:"description"`
	CreatedAt        *time.Time `json:"created_at"`
	UpdateAt         time.Time
}
