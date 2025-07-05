package services

import (
	"errors"
	"fmt"

	db "github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models/db"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser create a user record
func CreateUser(name string, email string, plainPassword string) (*db.User, error) {
	password, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("cannot generate hashed password")
	}

	user := db.NewUser(email, string(password), name, db.RoleUser)
	err = mgm.Coll(user).Create(user)
	if err != nil {
		return nil, errors.New("cannot create new user")
	}

	return user, nil
}

func CreateGoogleUser(name, email, profilePicUrl string) (*db.User, error) {
	user := db.NewGoogleUser(email, name, profilePicUrl)
	err := mgm.Coll(user).Create(user)
	if err != nil {
		return nil, errors.New("cannot create new user")
	}
	return user, nil
}

// FindUserById find user by id
func FindUserById(userId primitive.ObjectID) (*db.User, error) {
	user := &db.User{}
	err := mgm.Coll(user).FindByID(userId, user)
	if err != nil {
		return nil, errors.New("cannot find user")
	}

	return user, nil
}

// FindUserByEmail find user by email
func FindUserByEmail(email string) (*db.User, error) {
	user := &db.User{}
	err := mgm.Coll(user).First(bson.M{"email": email}, user)
	if err != nil {
		return nil, errors.New("cannot find user")
	}

	return user, nil
}

// CheckUserMail search user by email, return error if someone uses
func CheckUserMail(email string) error {
	user := &db.User{}
	userCollection := mgm.Coll(user)
	err := userCollection.First(bson.M{"email": email}, user)
	if err == nil {
		return errors.New("email is already in use")
	}

	return nil
}

func UpdateFCMToken(userId primitive.ObjectID, fcmToken string) error {
	user, err := FindUserById(userId)
	if err != nil {
		return err
	}

	user.FCMToken = fcmToken
	err = mgm.Coll(user).Update(user)
	if err != nil {
		return errors.New("cannot update fcm token")
	}

	return nil
}

func UpdateUserProfilePicture(userId primitive.ObjectID, profilePicUrl string) error {
	user, err := FindUserById(userId)
	if err != nil {
		return err
	}

	// Only update if the profile picture URL is different
	if user.ProfilePicUrl != profilePicUrl {
		user.ProfilePicUrl = profilePicUrl
		err = mgm.Coll(user).Update(user)
		if err != nil {
			return errors.New("cannot update profile picture")
		}
	}

	return nil
}

// UpdateUserProfile updates user's name only (email is not editable)
func UpdateUserProfile(userId primitive.ObjectID, name string) (*db.User, error) {
	user, err := FindUserById(userId)
	if err != nil {
		return nil, err
	}

	// Update name if provided
	if name != "" {
		user.Name = name
	}

	err = mgm.Coll(user).Update(user)
	if err != nil {
		return nil, errors.New("cannot update user profile")
	}

	return user, nil
}

// UpdateUserProfilePictureS3Key updates user's profile picture S3 key (for private bucket approach)
func UpdateUserProfilePictureS3Key(userId primitive.ObjectID, s3Key string) error {
	user, err := FindUserById(userId)
	if err != nil {
		return err
	}

	// Update S3 key
	user.ProfilePicS3Key = s3Key
	err = mgm.Coll(user).Update(user)
	if err != nil {
		return errors.New("cannot update profile picture S3 key")
	}

	return nil
}

// GetUserWithProfilePictureURL returns user with computed profile picture download URL
func GetUserWithProfilePictureURL(userId primitive.ObjectID, urlExpirationMinutes int) (*db.User, error) {
	user, err := FindUserById(userId)
	if err != nil {
		return nil, err
	}

	// Generate download URL if S3 key exists
	if user.ProfilePicS3Key != "" {
		downloadURL, err := GeneratePresignedDownloadURL(user.ProfilePicS3Key, urlExpirationMinutes)
		if err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: Failed to generate download URL for user %s: %s\n", userId.Hex(), err.Error())
		} else {
			user.ProfilePicUrl = downloadURL
		}
	}

	return user, nil
}
