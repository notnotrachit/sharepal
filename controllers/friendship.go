package controllers

import (
	"net/http"

	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/services"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SendFriendRequest godoc
// @Summary      Send Friend Request
// @Description  sends a friend request to another user by email
// @Tags         friends
// @Accept       json
// @Produce      json
// @Param        req  body      models.SendFriendRequestRequest true "Friend Request"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /friends/request [post]
// @Security     ApiKeyAuth
func SendFriendRequest(c *gin.Context) {
	var requestBody models.SendFriendRequestRequest
	_ = c.ShouldBindBodyWith(&requestBody, binding.JSON)

	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	err := services.SendFriendRequest(userId.(primitive.ObjectID), requestBody.Email)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Message = "Friend request sent successfully"
	response.SendResponse(c)
}

// RespondToFriendRequest godoc
// @Summary      Respond to Friend Request
// @Description  accepts or rejects a friend request
// @Tags         friends
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Friendship ID"
// @Param        req  body      models.RespondFriendRequestRequest true "Response Request"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /friends/request/{id}/respond [post]
// @Security     ApiKeyAuth
func RespondToFriendRequest(c *gin.Context) {
	var requestBody models.RespondFriendRequestRequest
	_ = c.ShouldBindBodyWith(&requestBody, binding.JSON)

	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	friendshipId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid friendship id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	err = services.RespondToFriendRequest(friendshipId, userId.(primitive.ObjectID), requestBody.Accept)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	message := "Friend request rejected"
	if requestBody.Accept {
		message = "Friend request accepted"
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Message = message
	response.SendResponse(c)
}

// GetFriends godoc
// @Summary      Get Friends
// @Description  gets all friends of the authenticated user
// @Tags         friends
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /friends [get]
// @Security     ApiKeyAuth
func GetFriends(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	friends, err := services.GetFriends(userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"friends": friends}
	response.SendResponse(c)
}

// GetPendingFriendRequests godoc
// @Summary      Get Pending Friend Requests
// @Description  gets all pending friend requests received by the authenticated user
// @Tags         friends
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /friends/requests/received [get]
// @Security     ApiKeyAuth
func GetPendingFriendRequests(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	requests, err := services.GetPendingFriendRequests(userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"requests": requests}
	response.SendResponse(c)
}

// GetSentFriendRequests godoc
// @Summary      Get Sent Friend Requests
// @Description  gets all pending friend requests sent by the authenticated user
// @Tags         friends
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /friends/requests/sent [get]
// @Security     ApiKeyAuth
func GetSentFriendRequests(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	requests, err := services.GetSentFriendRequests(userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"requests": requests}
	response.SendResponse(c)
}

// RemoveFriend godoc
// @Summary      Remove Friend
// @Description  removes a friend from the user's friend list
// @Tags         friends
// @Accept       json
// @Produce      json
// @Param        friendId  path      string  true  "Friend ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /friends/{friendId} [delete]
// @Security     ApiKeyAuth
func RemoveFriend(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	friendIdHex := c.Param("friendId")
	friendId, err := primitive.ObjectIDFromHex(friendIdHex)
	if err != nil {
		response.Message = "invalid friend id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	err = services.RemoveFriend(userId.(primitive.ObjectID), friendId)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Message = "Friend removed successfully"
	response.SendResponse(c)
}

// BlockUser godoc
// @Summary      Block User
// @Description  blocks a user
// @Tags         friends
// @Accept       json
// @Produce      json
// @Param        userId  path      string  true  "User ID to block"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /friends/block/{userId} [post]
// @Security     ApiKeyAuth
func BlockUser(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	targetUserIdHex := c.Param("userId")
	targetUserId, err := primitive.ObjectIDFromHex(targetUserIdHex)
	if err != nil {
		response.Message = "invalid user id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	err = services.BlockUser(userId.(primitive.ObjectID), targetUserId)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Message = "User blocked successfully"
	response.SendResponse(c)
}
