package services

import (
	"errors"
	"time"

	db "github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models/db"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SendFriendRequest(requesterID primitive.ObjectID, email string) error {
	// Find the addressee by email
	addressee, err := FindUserByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	// Can't send friend request to yourself
	if requesterID == addressee.ID {
		return errors.New("cannot send friend request to yourself")
	}

	// Check if friendship already exists
	existing := &db.Friendship{}
	err = mgm.Coll(existing).FindOne(mgm.Ctx(), bson.M{
		"$or": []bson.M{
			{
				"requester_id": requesterID,
				"addressee_id": addressee.ID,
			},
			{
				"requester_id": addressee.ID,
				"addressee_id": requesterID,
			},
		},
	}).Decode(existing)

	if err == nil {
		// Friendship exists
		switch existing.Status {
		case db.FriendshipAccepted:
			return errors.New("you are already friends")
		case db.FriendshipPending:
			return errors.New("friend request already sent")
		case db.FriendshipBlocked:
			return errors.New("cannot send friend request")
		}
	} else if err != mongo.ErrNoDocuments {
		return err
	}

	// Create new friend request
	friendship := db.NewFriendship(requesterID, addressee.ID)
	err = mgm.Coll(friendship).Create(friendship)

	return err
}

func RespondToFriendRequest(friendshipID primitive.ObjectID, addresseeID primitive.ObjectID, accept bool) error {
	friendship := &db.Friendship{}

	err := mgm.Coll(friendship).FindOne(mgm.Ctx(), bson.M{
		"_id":          friendshipID,
		"addressee_id": addresseeID,
		"status":       db.FriendshipPending,
	}).Decode(friendship)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("friend request not found")
		}
		return err
	}

	updateDoc := bson.M{}

	if accept {
		now := time.Now()
		updateDoc["status"] = db.FriendshipAccepted
		updateDoc["accepted_at"] = now
	} else {
		updateDoc["status"] = db.FriendshipRejected
	}

	_, err = mgm.Coll(friendship).UpdateOne(mgm.Ctx(), bson.M{"_id": friendshipID}, bson.M{
		"$set": updateDoc,
	})

	return err
}

func GetFriends(userID primitive.ObjectID) ([]*db.User, error) {
	var friendships []*db.Friendship

	// Find all accepted friendships where user is either requester or addressee
	err := mgm.Coll(&db.Friendship{}).SimpleFind(&friendships, bson.M{
		"$or": []bson.M{
			{"requester_id": userID},
			{"addressee_id": userID},
		},
		"status": db.FriendshipAccepted,
	})

	if err != nil {
		return nil, err
	}

	// Extract friend IDs
	var friendIDs []primitive.ObjectID
	for _, friendship := range friendships {
		if friendship.RequesterID == userID {
			friendIDs = append(friendIDs, friendship.AddresseeID)
		} else {
			friendIDs = append(friendIDs, friendship.RequesterID)
		}
	}

	// Get friend details
	var friends []*db.User
	if len(friendIDs) > 0 {
		err = mgm.Coll(&db.User{}).SimpleFind(&friends, bson.M{
			"_id": bson.M{"$in": friendIDs},
		})
	}

	return friends, err
}

func GetPendingFriendRequests(userID primitive.ObjectID) ([]*db.Friendship, error) {
	var friendships []*db.Friendship

	err := mgm.Coll(&db.Friendship{}).SimpleFind(&friendships, bson.M{
		"addressee_id": userID,
		"status":       db.FriendshipPending,
	})

	return friendships, err
}

func GetSentFriendRequests(userID primitive.ObjectID) ([]*db.Friendship, error) {
	var friendships []*db.Friendship

	err := mgm.Coll(&db.Friendship{}).SimpleFind(&friendships, bson.M{
		"requester_id": userID,
		"status":       db.FriendshipPending,
	})

	return friendships, err
}

func RemoveFriend(userID, friendID primitive.ObjectID) error {
	// Find the friendship
	friendship := &db.Friendship{}

	err := mgm.Coll(friendship).FindOne(mgm.Ctx(), bson.M{
		"$or": []bson.M{
			{
				"requester_id": userID,
				"addressee_id": friendID,
			},
			{
				"requester_id": friendID,
				"addressee_id": userID,
			},
		},
		"status": db.FriendshipAccepted,
	}).Decode(friendship)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("friendship not found")
		}
		return err
	}

	// Delete the friendship
	err = mgm.Coll(friendship).Delete(friendship)
	return err
}

func BlockUser(userID, targetUserID primitive.ObjectID) error {
	// Check if friendship exists
	friendship := &db.Friendship{}

	err := mgm.Coll(friendship).FindOne(mgm.Ctx(), bson.M{
		"$or": []bson.M{
			{
				"requester_id": userID,
				"addressee_id": targetUserID,
			},
			{
				"requester_id": targetUserID,
				"addressee_id": userID,
			},
		},
	}).Decode(friendship)

	if err == mongo.ErrNoDocuments {
		// Create new blocked relationship
		friendship = db.NewFriendship(userID, targetUserID)
		friendship.Status = db.FriendshipBlocked
		return mgm.Coll(friendship).Create(friendship)
	} else if err != nil {
		return err
	}

	// Update existing friendship to blocked
	_, err = mgm.Coll(friendship).UpdateOne(mgm.Ctx(), bson.M{"_id": friendship.ID}, bson.M{
		"$set": bson.M{"status": db.FriendshipBlocked},
	})

	return err
}
