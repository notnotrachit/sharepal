package db

import (
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Group struct {
	mgm.DefaultModel `bson:",inline"`
	Name             string               `json:"name" bson:"name"`
	Description      string               `json:"description" bson:"description"`
	CreatedBy        primitive.ObjectID   `json:"created_by" bson:"created_by"`
	Members          []primitive.ObjectID `json:"members" bson:"members"`
	IsActive         bool                 `json:"is_active" bson:"is_active"`
	Currency         string               `json:"currency" bson:"currency"` // USD, EUR, etc.
}

func NewGroup(name, description string, createdBy primitive.ObjectID, currency string) *Group {
	return &Group{
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		Members:     []primitive.ObjectID{createdBy}, // Creator is automatically a member
		IsActive:    true,
		Currency:    currency,
	}
}

func (model *Group) CollectionName() string {
	return "groups"
}
