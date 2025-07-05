package controllers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetPresignedUploadURL godoc
// @Summary      Get Presigned Upload URL
// @Description  Get a presigned URL for uploading profile picture to S3
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        req  body      models.PresignedURLRequest true "Presigned URL Request"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Failure      500  {object}  models.Response
// @Router       /media/presigned-upload-url [post]
// @Security     ApiKeyAuth
func GetPresignedUploadURL(c *gin.Context) {
	// Get user ID from context
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

	var request models.PresignedURLRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		models.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate file extension
	fileExt := strings.ToLower(filepath.Ext(request.FileName))
	if fileExt == "" {
		models.SendErrorResponse(c, http.StatusBadRequest, "File extension is required")
		return
	}

	// Generate presigned URL
	presignedData, err := services.GeneratePresignedUploadURL(userObjID, fileExt)
	if err != nil {
		models.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	models.SendSuccessResponse(c, "Presigned URL generated successfully", map[string]any{
		"upload_url": presignedData.UploadURL,
		"s3_key":     presignedData.S3Key,
		"expires_at": presignedData.ExpiresAt,
	})
}

// ConfirmProfilePictureUpload godoc
// @Summary      Confirm Profile Picture Upload
// @Description  Confirm that profile picture was uploaded to S3 and update user record
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        req  body      models.ConfirmUploadRequest true "Confirm Upload Request"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Failure      500  {object}  models.Response
// @Router       /media/confirm-upload [post]
// @Security     ApiKeyAuth
func ConfirmProfilePictureUpload(c *gin.Context) {
	// Get user ID from context
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

	var request models.ConfirmUploadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		models.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate that the upload was successful
	err := services.ValidateS3Upload(request.S3Key)
	if err != nil {
		models.SendErrorResponse(c, http.StatusBadRequest, "Upload validation failed: file not found in S3")
		return
	}

	// Get current user to cleanup old profile picture
	user, err := services.FindUserById(userObjID)
	if err != nil {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to find user")
		return
	}

	// Delete old profile picture from S3 if exists
	if user.ProfilePicUrl != "" {
		// Extract S3 key from old URL if it's an S3 URL
		if strings.Contains(user.ProfilePicUrl, ".amazonaws.com/") {
			parts := strings.Split(user.ProfilePicUrl, ".amazonaws.com/")
			if len(parts) == 2 {
				oldS3Key := parts[1]
				_ = services.DeleteS3Object(oldS3Key) // Don't fail if cleanup fails
			}
		}
	}

	// Update user's profile picture S3 key (for private bucket)
	err = services.UpdateUserProfilePictureS3Key(userObjID, request.S3Key)
	if err != nil {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update user profile picture")
		return
	}

	// Generate presigned download URL for immediate response
	downloadURL, err := services.GeneratePresignedDownloadURL(request.S3Key, 60) // 60 minutes
	if err != nil {
		// Still return success since the upload worked, just no download URL
		models.SendSuccessResponse(c, "Profile picture updated successfully", map[string]any{
			"profile_pic_url": "",
		})
		return
	}

	models.SendSuccessResponse(c, "Profile picture updated successfully", map[string]any{
		"profile_pic_url": downloadURL,
	})
}

// Note: GetMedia endpoint is no longer needed as files are served directly from S3

// DeleteProfilePicture godoc
// @Summary      Delete Profile Picture
// @Description  Delete the current user's profile picture from S3
// @Tags         media
// @Produce      json
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Failure      500  {object}  models.Response
// @Router       /media/profile-picture [delete]
// @Security     ApiKeyAuth
func DeleteProfilePicture(c *gin.Context) {
	// Get user ID from context
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

	// Get current user to check if they have a profile picture
	user, err := services.FindUserById(userObjID)
	if err != nil {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to find user")
		return
	}

	if user.ProfilePicUrl == "" {
		models.SendErrorResponse(c, http.StatusBadRequest, "No profile picture to delete")
		return
	}

	// Delete the file from S3 if it's an S3 URL
	if strings.Contains(user.ProfilePicUrl, ".amazonaws.com/") {
		parts := strings.Split(user.ProfilePicUrl, ".amazonaws.com/")
		if len(parts) == 2 {
			s3Key := parts[1]
			err = services.DeleteS3Object(s3Key)
			if err != nil {
				// Log the error but don't fail the request since we still want to update the user
				models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete file from S3")
				return
			}
		}
	}

	// Update user's profile picture S3 key to empty
	err = services.UpdateUserProfilePictureS3Key(userObjID, "")
	if err != nil {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update user profile picture")
		return
	}

	models.SendSuccessResponse(c, "Profile picture deleted successfully", nil)
}