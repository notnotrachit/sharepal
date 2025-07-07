package db

import (
	"time"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GroupBalance maintains running balance for each user in a group
type GroupBalance struct {
	mgm.DefaultModel `bson:",inline"`

	GroupID  primitive.ObjectID `json:"group_id" bson:"group_id"`
	UserID   primitive.ObjectID `json:"user_id" bson:"user_id"`
	UserName string             `json:"user_name" bson:"user_name"` // Denormalized for performance

	// Balance tracking
	Balance   float64 `json:"balance" bson:"balance"`       // Current net balance (positive = owed money, negative = owes money)
	TotalPaid float64 `json:"total_paid" bson:"total_paid"` // Total amount this user has paid
	TotalOwed float64 `json:"total_owed" bson:"total_owed"` // Total amount this user owes

	// Metadata
	Currency          string             `json:"currency" bson:"currency"`
	LastTransactionID primitive.ObjectID `json:"last_transaction_id" bson:"last_transaction_id"`
	LastUpdated       time.Time          `json:"last_updated" bson:"last_updated"`
	Version           int64              `json:"version" bson:"version"` // For optimistic locking
}

func NewGroupBalance(groupID, userID primitive.ObjectID, userName, currency string) *GroupBalance {
	return &GroupBalance{
		GroupID:     groupID,
		UserID:      userID,
		UserName:    userName,
		Balance:     0.0,
		TotalPaid:   0.0,
		TotalOwed:   0.0,
		Currency:    currency,
		LastUpdated: time.Now(),
		Version:     1,
	}
}

func (model *GroupBalance) CollectionName() string {
	return "group_balances"
}

// UpdateBalance applies a transaction to the balance
func (gb *GroupBalance) UpdateBalance(amountPaid, amountOwed float64, transactionID primitive.ObjectID) {
	gb.TotalPaid += amountPaid
	gb.TotalOwed += amountOwed
	gb.Balance = gb.TotalPaid - gb.TotalOwed
	gb.LastTransactionID = transactionID
	gb.LastUpdated = time.Now()
	gb.Version++
}
