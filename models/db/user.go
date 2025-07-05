package models

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/kamva/mgm/v3"
)

const (
	RoleUser = "user"
)

type User struct {
	mgm.DefaultModel `bson:",inline"`
	Email            string `json:"email" bson:"email"`
	Password         string `json:"-" bson:"password"`
	Name             string `json:"name" bson:"name"`
	Role             string `json:"role" bson:"role"`
	MailVerified     bool   `json:"mail_verified" bson:"mail_verified"`
	FCMToken         string `json:"fcm_token" bson:"fcm_token"`
	ProfilePicS3Key  string `json:"-" bson:"profile_pic_s3_key"` // Store S3 key privately for uploaded images
	ProfilePicUrl    string `json:"profile_pic_url,omitempty" bson:"profile_pic_url,omitempty"` // Store external URLs (Google, etc.) or computed S3 URLs
	ProfilePicType   string `json:"-" bson:"profile_pic_type"` // "s3", "external", or empty
}

type UserClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Type  string `json:"type"`
}

func NewUser(email string, password string, name string, role string) *User {
	return &User{
		Email:    email,
		Password: password,
		Name:     name,
		Role:     role,
		MailVerified: false,
		FCMToken:     "",
	}
}

func NewGoogleUser(email string, name string, profilePicUrl string) *User {
	return &User{
		Email:           email,
		Name:            name,
		ProfilePicUrl:   profilePicUrl,
		ProfilePicType:  "external", // Mark as external URL (Google)
		Role:            RoleUser,
		MailVerified:    true,
	}
}

func (model *User) CollectionName() string {
	return "users"
}

// You can override Collection functions or CRUD hooks
// https://github.com/Kamva/mgm#a-models-hooks
// https://github.com/Kamva/mgm#collections
