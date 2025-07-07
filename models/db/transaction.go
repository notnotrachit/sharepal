package db

import (
	"time"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionType string

const (
	// Transaction Types
	TransactionTypeExpense    TransactionType = "expense"
	TransactionTypeSettlement TransactionType = "settlement"
	TransactionTypeRefund     TransactionType = "refund"
	TransactionTypeAdjustment TransactionType = "adjustment"
)

type SplitType string

const (
	// Split Types for expenses
	SplitTypeEqual      SplitType = "equal"
	SplitTypeExact      SplitType = "exact"
	SplitTypePercentage SplitType = "percentage"
)

// TransactionPayer represents who actually paid money
type TransactionPayer struct {
	UserID   primitive.ObjectID `json:"user_id" bson:"user_id"`
	UserName string             `json:"user_name" bson:"user_name"`
	Amount   float64            `json:"amount" bson:"amount"` // Amount they paid
}

// TransactionSplit represents how the expense should be divided
type TransactionSplit struct {
	UserID   primitive.ObjectID `json:"user_id" bson:"user_id"`
	UserName string             `json:"user_name" bson:"user_name"`
	Amount   float64            `json:"amount" bson:"amount"` // Amount they owe
}

// TransactionParticipant represents a user's net involvement (for backward compatibility)
type TransactionParticipant struct {
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	UserName  string             `json:"user_name" bson:"user_name"`
	Amount    float64            `json:"amount" bson:"amount"`         // Net amount (paid - owed)
	ShareType string             `json:"share_type" bson:"share_type"` // "payer", "split", "both"
}

// Transaction replaces both Expense and Settlement models
type Transaction struct {
	mgm.DefaultModel `bson:",inline"`

	// Basic Info
	GroupID     primitive.ObjectID `json:"group_id" bson:"group_id"`
	Type        TransactionType    `json:"type" bson:"type"`
	Description string             `json:"description" bson:"description"`
	Amount      float64            `json:"amount" bson:"amount"` // Total transaction amount
	Currency    string             `json:"currency" bson:"currency"`
	Date        time.Time          `json:"date" bson:"date"`

	// New structure for expenses
	Payers []TransactionPayer `json:"payers,omitempty" bson:"payers,omitempty"` // Who paid money
	Splits []TransactionSplit `json:"splits,omitempty" bson:"splits,omitempty"` // How expense is divided

	// Legacy participants field (for backward compatibility and settlements)
	Participants []TransactionParticipant `json:"participants" bson:"participants"`

	// Expense-specific fields (only for expense type)
	Category  string    `json:"category,omitempty" bson:"category,omitempty"`
	SplitType SplitType `json:"split_type,omitempty" bson:"split_type,omitempty"`
	Receipt   string    `json:"receipt,omitempty" bson:"receipt,omitempty"`

	// Settlement-specific fields (only for settlement type)
	SettledAt        *time.Time `json:"settled_at,omitempty" bson:"settled_at,omitempty"`
	SettlementMethod string     `json:"settlement_method,omitempty" bson:"settlement_method,omitempty"`
	ProofOfPayment   string     `json:"proof_of_payment,omitempty" bson:"proof_of_payment,omitempty"`

	// Common fields
	Notes       string             `json:"notes" bson:"notes"`
	IsCompleted bool               `json:"is_completed" bson:"is_completed"`
	CreatedBy   primitive.ObjectID `json:"created_by" bson:"created_by"`

	// Audit trail
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	UpdatedBy primitive.ObjectID `json:"updated_by,omitempty" bson:"updated_by,omitempty"`
}

func NewExpenseTransaction(groupID primitive.ObjectID, description string, amount float64, currency string, paidBy primitive.ObjectID, splitType SplitType, category string) *Transaction {
	return &Transaction{
		GroupID:      groupID,
		Type:         TransactionTypeExpense,
		Description:  description,
		Amount:       amount,
		Currency:     currency,
		Date:         time.Now(),
		Category:     category,
		SplitType:    splitType,
		Participants: []TransactionParticipant{},
		IsCompleted:  false,
		CreatedBy:    paidBy,
		UpdatedAt:    time.Now(),
	}
}

func NewSettlementTransaction(groupID, payerID, payeeID primitive.ObjectID, amount float64, currency string) *Transaction {
	return &Transaction{
		GroupID:     groupID,
		Type:        TransactionTypeSettlement,
		Description: "Settlement",
		Amount:      amount,
		Currency:    currency,
		Date:        time.Now(),
		Participants: []TransactionParticipant{
			{
				UserID:    payerID,
				Amount:    amount, // Positive = they're paying
				ShareType: "payer",
			},
			{
				UserID:    payeeID,
				Amount:    -amount, // Negative = they're receiving
				ShareType: "payee",
			},
		},
		IsCompleted: false,
		CreatedBy:   payerID,
		UpdatedAt:   time.Now(),
	}
}

func (model *Transaction) CollectionName() string {
	return "transactions"
}
