package models

type Finance struct {
	ID             uint    `gorm:"primaryKey" json:"id"`
	AccountTitle   string  `json:"account_title"`
	InitialBalance float64 `json:"initial_balance"`
	AccountNumber  string  `json:"account_number"`
	BranchCode     string  `json:"branch_code"`
	BankBranch     string  `json:"bank_branch"`
}
