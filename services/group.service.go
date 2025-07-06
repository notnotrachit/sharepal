package services

import (
	"errors"

	db "github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models/db"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateGroup(name, description, currency string, createdBy primitive.ObjectID, memberIDs []string) (*db.Group, error) {
	group := db.NewGroup(name, description, createdBy, currency)

	// Add additional members if provided
	for _, memberIDStr := range memberIDs {
		memberID, err := primitive.ObjectIDFromHex(memberIDStr)
		if err != nil {
			continue // Skip invalid IDs
		}

		// Check if user exists
		if _, err := FindUserById(memberID); err != nil {
			continue // Skip non-existent users
		}

		// Don't add creator twice
		if memberID != createdBy {
			group.Members = append(group.Members, memberID)
		}
	}

	err := mgm.Coll(group).Create(group)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func GetUserGroups(userID primitive.ObjectID, page, limit int) ([]*db.Group, error) {
	var groups []*db.Group

	findOptions := options.Find().
		SetSkip(int64(page * limit)).
		SetLimit(int64(limit + 1)) // +1 to check if there are more

	err := mgm.Coll(&db.Group{}).SimpleFind(&groups, bson.M{
		"members":   userID,
		"is_active": true,
	}, findOptions)

	return groups, err
}

func GetGroupById(groupID, userID primitive.ObjectID) (*db.Group, error) {
	group := &db.Group{}

	err := mgm.Coll(group).FindOne(mgm.Ctx(), bson.M{
		"_id":       groupID,
		"members":   userID,
		"is_active": true,
	}).Decode(group)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("group not found or access denied")
		}
		return nil, err
	}

	return group, nil
}

func AddMemberToGroup(groupID, userID, newMemberID primitive.ObjectID) error {
	// Check if user is group member
	group, err := GetGroupById(groupID, userID)
	if err != nil {
		return err
	}

	// Check if new member exists
	if _, err := FindUserById(newMemberID); err != nil {
		return errors.New("user not found")
	}

	// Check if already a member
	for _, memberID := range group.Members {
		if memberID == newMemberID {
			return errors.New("user is already a member")
		}
	}

	// Add member
	_, err = mgm.Coll(group).UpdateOne(mgm.Ctx(), bson.M{"_id": groupID}, bson.M{
		"$push": bson.M{"members": newMemberID},
	})

	return err
}

func RemoveMemberFromGroup(groupID, userID, memberToRemoveID primitive.ObjectID) error {
	// Check if user is group member or creator
	group, err := GetGroupById(groupID, userID)
	if err != nil {
		return err
	}

	// Only creator can remove members
	if group.CreatedBy != userID {
		return errors.New("only group creator can remove members")
	}

	// Cannot remove creator
	if memberToRemoveID == group.CreatedBy {
		return errors.New("cannot remove group creator")
	}

	// Remove member
	_, err = mgm.Coll(group).UpdateOne(mgm.Ctx(), bson.M{"_id": groupID}, bson.M{
		"$pull": bson.M{"members": memberToRemoveID},
	})

	return err
}

func DeleteGroup(groupID, userID primitive.ObjectID) error {
	// Check if user is group creator
	group, err := GetGroupById(groupID, userID)
	if err != nil {
		return err
	}

	if group.CreatedBy != userID {
		return errors.New("only group creator can delete the group")
	}

	// Soft delete by setting is_active to false
	_, err = mgm.Coll(group).UpdateOne(mgm.Ctx(), bson.M{"_id": groupID}, bson.M{
		"$set": bson.M{"is_active": false},
	})

	return err
}

func GetGroupMembers(groupID, userID primitive.ObjectID) ([]*db.User, error) {
	// Check if user is group member
	group, err := GetGroupById(groupID, userID)
	if err != nil {
		return nil, err
	}

	var users []*db.User
	err = mgm.Coll(&db.User{}).SimpleFind(&users, bson.M{
		"_id": bson.M{"$in": group.Members},
	})

	return users, err
}

func UpdateGroup(groupID, userID primitive.ObjectID, name, description, currency string) (*db.Group, error) {
	group, err := GetGroupById(groupID, userID)
	if err != nil {
		return nil, err
	}

	if group.CreatedBy != userID {
		return nil, errors.New("only group creator can update group details")
	}

	updateDoc := bson.M{}
	
	if name != "" {
		updateDoc["name"] = name
	}
	if description != "" {
		updateDoc["description"] = description
	}
	if currency != "" {
		updateDoc["currency"] = currency
	}

	if len(updateDoc) == 0 {
		return group, nil
	}

	_, err = mgm.Coll(group).UpdateOne(mgm.Ctx(), bson.M{"_id": groupID}, bson.M{
		"$set": updateDoc,
	})
	if err != nil {
		return nil, err
	}

	return GetGroupById(groupID, userID)
}
