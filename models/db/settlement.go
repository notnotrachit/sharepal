package models

import (
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type SettlementStatus string

const (
	SettlementPending   SettlementStatus = "pending"
	SettlementCompleted SettlementStatus = "completed"
	SettlementCancelled SettlementStatus = "cancelled"
)

type Settlement struct {
	mgm.DefaultModel `bson:",inline"`
	GroupID          primitive.ObjectID `json:"group_id" bson:"group_id"`
	PayerID          primitive.ObjectID `json:"payer_id" bson:"payer_id"`     // Who needs to pay
	PayeeID          primitive.ObjectID `json:"payee_id" bson:"payee_id"`     // Who should receive
	Amount           float64            `json:"amount" bson:"amount"`
	Currency         string             `json:"currency" bson:"currency"`
	Status           SettlementStatus   `json:"status" bson:"status"`
	SettledAt        *time.Time         `json:"settled_at,omitempty" bson:"settled_at,omitempty"`
	Notes            string             `json:"notes" bson:"notes"`
	ExpenseIDs       []primitive.ObjectID `json:"expense_ids" bson:"expense_ids"` // Related expenses
}

func NewSettlement(groupID, payerID, payeeID primitive.ObjectID, amount float64, currency string) *Settlement {
	return &Settlement{
		GroupID:  groupID,
		PayerID:  payerID,
		PayeeID:  payeeID,
		Amount:   amount,
		Currency: currency,
		Status:   SettlementPending,
	}
}

func (model *Settlement) CollectionName() string {
	return "settlements"
}
