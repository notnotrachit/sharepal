package models

import (
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type SplitType string

const (
	SplitTypeEqual      SplitType = "equal"
	SplitTypeExact      SplitType = "exact"
	SplitTypePercentage SplitType = "percentage"
)

type ExpenseSplit struct {
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
	Amount float64            `json:"amount" bson:"amount"`
}

type Expense struct {
	mgm.DefaultModel `bson:",inline"`
	GroupID          primitive.ObjectID `json:"group_id" bson:"group_id"`
	Description      string             `json:"description" bson:"description"`
	Amount           float64            `json:"amount" bson:"amount"`
	Currency         string             `json:"currency" bson:"currency"`
	PaidBy           primitive.ObjectID `json:"paid_by" bson:"paid_by"`
	SplitType        SplitType          `json:"split_type" bson:"split_type"`
	Splits           []ExpenseSplit     `json:"splits" bson:"splits"`
	Category         string             `json:"category" bson:"category"`
	Date             time.Time          `json:"date" bson:"date"`
	Receipt          string             `json:"receipt" bson:"receipt"` // URL to receipt image
	IsSettled        bool               `json:"is_settled" bson:"is_settled"`
	Notes            string             `json:"notes" bson:"notes"`
}

func NewExpense(groupID primitive.ObjectID, description string, amount float64, currency string, paidBy primitive.ObjectID, splitType SplitType, category string) *Expense {
	return &Expense{
		GroupID:     groupID,
		Description: description,
		Amount:      amount,
		Currency:    currency,
		PaidBy:      paidBy,
		SplitType:   splitType,
		Splits:      []ExpenseSplit{},
		Category:    category,
		Date:        time.Now(),
		IsSettled:   false,
	}
}

func (model *Expense) CollectionName() string {
	return "expenses"
}
