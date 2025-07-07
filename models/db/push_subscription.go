package db

import (
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PushSubscription struct {
	mgm.DefaultModel `bson:",inline"`
	UserID           primitive.ObjectID `json:"user_id" bson:"user_id"`
	Endpoint         string             `json:"endpoint" bson:"endpoint"`
	P256dh           string             `json:"p256dh" bson:"p256dh"`
	Auth             string             `json:"auth" bson:"auth"`
}

func NewPushSubscription(userID primitive.ObjectID, endpoint, p256dh, auth string) *PushSubscription {
	return &PushSubscription{
		UserID:   userID,
		Endpoint: endpoint,
		P256dh:   p256dh,
		Auth:     auth,
	}
}