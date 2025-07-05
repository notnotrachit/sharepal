package controllers

import (
	"fmt"
	"net/http"

	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UpdateFCMToken(c *gin.Context) {
	var request models.UpdateFCMTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		models.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		models.SendErrorResponse(c, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userObjID, ok := userId.(primitive.ObjectID)
	if !ok {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID in context")
		return
	}

	fmt.Printf("Updating FCM token for user ID: %s\n", userObjID.Hex())
	fmt.Printf("New FCM token: %s\n", request.FCMToken)

	err := services.UpdateFCMToken(userObjID, request.FCMToken)
	if err != nil {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update FCM token")
		return
	}

	models.SendSuccessResponse(c, "FCM token updated successfully", nil)
}

// UpdateProfile godoc
// @Summary      Update User Profile
// @Description  Update the current user's profile information (name only)
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        req  body      models.UpdateProfileRequest true "Update Profile Request"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Failure      500  {object}  models.Response
// @Router       /user/profile [put]
// @Security     ApiKeyAuth
func UpdateProfile(c *gin.Context) {
	var request models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		models.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := request.Validate(); err != nil {
		models.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		models.SendErrorResponse(c, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	userObjID, ok := userId.(primitive.ObjectID)
	if !ok {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID in context")
		return
	}

	// Update user profile (name only)
	updatedUser, err := services.UpdateUserProfile(userObjID, request.Name)
	if err != nil {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	userWithProfilePic, err := services.GetUserWithProfilePictureURL(updatedUser.ID, 60)
	if err != nil {
		userWithProfilePic = updatedUser
	}

	models.SendSuccessResponse(c, "Profile updated successfully", map[string]any{
		"user": userWithProfilePic,
	})
}
