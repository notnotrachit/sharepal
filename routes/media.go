package routes

import (
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/controllers"
	"github.com/gin-gonic/gin"
)

func MediaRoute(router *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	media := router.Group("/media", handlers...)
	{
		// Presigned URL for S3 uploads
		media.POST("/presigned-upload-url", controllers.GetPresignedUploadURL)
		
		// Confirm upload completion
		media.POST("/confirm-upload", controllers.ConfirmProfilePictureUpload)
		
		// Delete profile picture
		media.DELETE("/profile-picture", controllers.DeleteProfilePicture)
	}
}