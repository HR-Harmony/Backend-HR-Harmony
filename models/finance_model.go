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

type AddDeposit struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	FinanceID       uint      `json:"finance_id"`
	AccountTitle    string    `json:"account_title"`
	Amount          float64   `json:"amount"`
	Date            string    `json:"date"` // Format: yyyy-mm-dd
	CategoryID      uint      `json:"category_id"`
	DepositCategory string    `json:"deposit_category"`
	Payer           string    `json:"payer"`
	PaymentMethod   string    `json:"payment_method"`
	Ref             string    `json:"ref"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

type ExpenseCategory struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ExpenseCategory string    `json:"expense_category"`
	CreatedAt       time.Time `json:"created_at"`
}

type AddExpense struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	FinanceID         uint      `json:"finance_id"`
	AccountTitle      string    `json:"account_title"`
	Amount            float64   `json:"amount"`
	Date              string    `json:"date"`
	ExpenseCategoryID uint      `json:"expense_category_id"`
	ExpenseCategory   string    `json:"expense_category"`
	Payer             string    `json:"payer"`
	PaymentMethod     string    `json:"payment_method"`
	Ref               string    `json:"ref"`
	Description       string    `json:"description"`
	CreatedAt         time.Time `json:"created_at"`
}
