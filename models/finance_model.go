package models

import "time"

type Finance struct {
	ID             uint    `gorm:"primaryKey" json:"id"`
	AccountTitle   string  `json:"account_title"`
	InitialBalance float64 `json:"initial_balance"`
	AccountNumber  string  `json:"account_number"`
	BranchCode     string  `json:"branch_code"`
	BankBranch     string  `json:"bank_branch"`
}

type DepositCategory struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	DepositCategory string    `json:"deposit_category"`
	CreatedAt       time.Time `json:"created_at"`
}
