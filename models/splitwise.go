package models

import "time"

type SplitwiseExpense struct {
	ID          int64         `json:"id"`
	Description string        `json:"description"`
	Cost        string        `json:"cost"`
	Date        time.Time     `json:"date"`
	Currency    string        `json:"currency_code"`
	Repayments  []Repayment   `json:"repayments"`
	DeletedAt   *time.Time    `json:"deleted_at"`
	DeletedBy   *User         `json:"deleted_by"`
	Users       []ExpenseUser `json:"users"`
}

type Repayment struct {
	From   int64  `json:"from"`
	To     int64  `json:"to"`
	Amount string `json:"amount"`
}

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ExpenseUser struct {
	User       User   `json:"user"`
	UserID     int64  `json:"user_id"`
	PaidShare  string `json:"paid_share"`
	OwedShare  string `json:"owed_share"`
	NetBalance string `json:"net_balance"`
}

type SplitwiseComment struct {
	ID           int64      `json:"id"`
	Content      string     `json:"content"`
	CommentType  string     `json:"comment_type"`
	RelationType string     `json:"relation_type"`
	RelationID   int64      `json:"relation_id"`
	CreatedAt    time.Time  `json:"created_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
	User         User       `json:"user"`
}
