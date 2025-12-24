package models

// UpdateAction - only the expense ID and what changed
type UpdateAction struct {
	ExpenseID       int64
	Expense         SplitwiseExpense
	ChangedFields   []string // e.g., ["amount", "description", "repayments"]
	OldHash         string   // Previous snapshot hash
	NewHash         string   // Current snapshot hash
	LunchMoneyTxnID string   // Existing LM transaction to update
}

// DeleteAction - just identifiers needed to delete
type DeleteAction struct {
	ExpenseID       int64
	LunchMoneyTxnID string // LM transaction ID to delete
}
