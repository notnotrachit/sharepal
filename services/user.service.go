package services

import (
	"errors"
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

func GetPushSubscriptionsByUserID(userID primitive.ObjectID) ([]*db.PushSubscription, error) {
	var subscriptions []*db.PushSubscription
	err := mgm.Coll(&db.PushSubscription{}).SimpleFind(&subscriptions, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	return subscriptions, nil
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

	// Update S3 key and clear external URL
	user.ProfilePicS3Key = s3Key
	user.ProfilePicUrl = "" // Clear any external URL
	user.ProfilePicType = "s3" // Mark as S3-hosted
	err = mgm.Coll(user).Update(user)
	if err != nil {
		return errors.New("cannot update profile picture S3 key")
	}

	return nil
}

// UpdateUserProfilePictureExternalURL updates user's profile picture with external URL (Google, etc.)
func UpdateUserProfilePictureExternalURL(userId primitive.ObjectID, externalUrl string) error {
	user, err := FindUserById(userId)
	if err != nil {
		return err
	}

	// Update external URL and clear S3 key
	user.ProfilePicUrl = externalUrl
	user.ProfilePicS3Key = "" // Clear any S3 key
	user.ProfilePicType = "external" // Mark as external URL
	err = mgm.Coll(user).Update(user)
	if err != nil {
		return errors.New("cannot update profile picture external URL")
	}

	return nil
}

// GetUserWithProfilePictureURL returns user with computed profile picture download URL
func GetUserWithProfilePictureURL(userId primitive.ObjectID, urlExpirationMinutes int) (*db.User, error) {
	user, err := FindUserById(userId)
	if err != nil {
		return nil, err
	}

	// Handle profile picture based on type
	switch user.ProfilePicType {
	case "s3":
		// Generate presigned download URL for S3-hosted images
		if user.ProfilePicS3Key != "" {
			downloadURL, err := GeneratePresignedDownloadURL(user.ProfilePicS3Key, urlExpirationMinutes)
			if err != nil {
				// Log error but don't fail the request
				user.ProfilePicUrl = "" // Clear invalid URL
			} else {
				user.ProfilePicUrl = downloadURL
			}
		}
	case "external":
		// External URLs (Google, etc.) are used as-is, already stored in ProfilePicUrl
		// No processing needed
	default:
		// Legacy support: if no type is set, check if we have S3 key or external URL
		if user.ProfilePicS3Key != "" {
			// Assume it's S3 and generate presigned URL
			downloadURL, err := GeneratePresignedDownloadURL(user.ProfilePicS3Key, urlExpirationMinutes)
			if err != nil {
				user.ProfilePicUrl = ""
			} else {
				user.ProfilePicUrl = downloadURL
			}
		}
		// If ProfilePicUrl is already set and no S3 key, assume it's external
	}

	return user, nil
}
