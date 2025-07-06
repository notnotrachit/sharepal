package controllers

import (
	"net/http"
	"strconv"

	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/services"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateGroup godoc
// @Summary      Create Group
// @Description  creates a new group for expense sharing
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        req  body      models.CreateGroupRequest true "Group Request"
// @Success      201  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups [post]
// @Security     ApiKeyAuth
func CreateGroup(c *gin.Context) {
	var requestBody models.CreateGroupRequest
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

	group, err := services.CreateGroup(requestBody.Name, requestBody.Description, requestBody.Currency, userId.(primitive.ObjectID), requestBody.MemberIDs)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusCreated
	response.Success = true
	response.Data = gin.H{"group": group}
	response.SendResponse(c)
}

// GetUserGroups godoc
// @Summary      Get User Groups
// @Description  gets user groups with pagination
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        page  query    string  false  "Switch page by 'page'"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups [get]
// @Security     ApiKeyAuth
func GetUserGroups(c *gin.Context) {
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

	pageQuery := c.DefaultQuery("page", "0")
	page, _ := strconv.Atoi(pageQuery)
	limit := 10

	groups, err := services.GetUserGroups(userId.(primitive.ObjectID), page, limit)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	hasPrev := page > 0
	hasNext := len(groups) > limit
	if hasNext {
		groups = groups[:limit]
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"groups": groups, "prev": hasPrev, "next": hasNext}
	response.SendResponse(c)
}

// GetGroupById godoc
// @Summary      Get Group
// @Description  get group by id
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{id} [get]
// @Security     ApiKeyAuth
func GetGroupById(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	groupId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	group, err := services.GetGroupById(groupId, userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"group": group}
	response.SendResponse(c)
}

// AddMemberToGroup godoc
// @Summary      Add Member to Group
// @Description  adds a member to a group
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Param        req  body      models.AddMemberToGroupRequest true "Add Member Request"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{id}/members [post]
// @Security     ApiKeyAuth
func AddMemberToGroup(c *gin.Context) {
	var requestBody models.AddMemberToGroupRequest
	_ = c.ShouldBindBodyWith(&requestBody, binding.JSON)

	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	groupId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	newMemberId, err := primitive.ObjectIDFromHex(requestBody.UserID)
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

	err = services.AddMemberToGroup(groupId, userId.(primitive.ObjectID), newMemberId)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Message = "Member added successfully"
	response.SendResponse(c)
}

// RemoveMemberFromGroup godoc
// @Summary      Remove Member from Group
// @Description  removes a member from a group
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        id        path      string  true  "Group ID"
// @Param        memberId  path      string  true  "Member ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{id}/members/{memberId} [delete]
// @Security     ApiKeyAuth
func RemoveMemberFromGroup(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	groupId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	memberIdHex := c.Param("memberId")
	memberId, err := primitive.ObjectIDFromHex(memberIdHex)
	if err != nil {
		response.Message = "invalid member id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	err = services.RemoveMemberFromGroup(groupId, userId.(primitive.ObjectID), memberId)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Message = "Member removed successfully"
	response.SendResponse(c)
}

// GetGroupMembers godoc
// @Summary      Get Group Members
// @Description  gets all members of a group
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{id}/members [get]
// @Security     ApiKeyAuth
func GetGroupMembers(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	groupId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	members, err := services.GetGroupMembers(groupId, userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"members": members}
	response.SendResponse(c)
}

// DeleteGroup godoc
// @Summary      Delete Group
// @Description  deletes a group (soft delete)
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{id} [delete]
// @Security     ApiKeyAuth
func DeleteGroup(c *gin.Context) {
	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	groupId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	err = services.DeleteGroup(groupId, userId.(primitive.ObjectID))
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Message = "Group deleted successfully"
	response.SendResponse(c)
}

// UpdateGroup godoc
// @Summary      Update Group
// @Description  updates group details (name, description, currency)
// @Tags         groups
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Group ID"
// @Param        req  body      models.UpdateGroupRequest true "Update Group Request"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Router       /groups/{id} [put]
// @Security     ApiKeyAuth
func UpdateGroup(c *gin.Context) {
	var requestBody models.UpdateGroupRequest
	_ = c.ShouldBindBodyWith(&requestBody, binding.JSON)

	response := &models.Response{
		StatusCode: http.StatusBadRequest,
		Success:    false,
	}

	idHex := c.Param("id")
	groupId, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		response.Message = "invalid group id"
		response.SendResponse(c)
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		response.Message = "cannot get user"
		response.SendResponse(c)
		return
	}

	updatedGroup, err := services.UpdateGroup(groupId, userId.(primitive.ObjectID), requestBody.Name, requestBody.Description, requestBody.Currency)
	if err != nil {
		response.Message = err.Error()
		response.SendResponse(c)
		return
	}

	response.StatusCode = http.StatusOK
	response.Success = true
	response.Data = gin.H{"group": updatedGroup}
	response.Message = "Group updated successfully"
	response.SendResponse(c)
}
