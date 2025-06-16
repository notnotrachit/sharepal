package models

import (
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type FriendshipStatus string

const (
	FriendshipPending  FriendshipStatus = "pending"
	FriendshipAccepted FriendshipStatus = "accepted"
	FriendshipRejected FriendshipStatus = "rejected"
	FriendshipBlocked  FriendshipStatus = "blocked"
)

type Friendship struct {
	mgm.DefaultModel `bson:",inline"`
	RequesterID      primitive.ObjectID `json:"requester_id" bson:"requester_id"`
	AddresseeID      primitive.ObjectID `json:"addressee_id" bson:"addressee_id"`
	Status           FriendshipStatus   `json:"status" bson:"status"`
	RequestedAt      time.Time          `json:"requested_at" bson:"requested_at"`
	AcceptedAt       *time.Time         `json:"accepted_at,omitempty" bson:"accepted_at,omitempty"`
}

func NewFriendship(requesterID, addresseeID primitive.ObjectID) *Friendship {
	return &Friendship{
		RequesterID: requesterID,
		AddresseeID: addresseeID,
		Status:      FriendshipPending,
		RequestedAt: time.Now(),
	}
}

func (model *Friendship) CollectionName() string {
	return "friendships"
}
