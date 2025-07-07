package controllers

import (
	"log"
	"net/http"

	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models"
	db "github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models/db"
	"github.com/ebubekiryigit/golang-mongodb-rest-api-starter/services"
	"github.com/gin-gonic/gin"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// RegisterPushSubscription handles the registration of a new push subscription
// @Summary Register Push Subscription
// @Description Registers a new push subscription for a user
// @Tags Push
// @Accept json
// @Produce json
// @Param subscription body models.PushSubscriptionRequest true "Push Subscription Request"
// @Security ApiKeyAuth
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /api/register-push [post]
func RegisterPushSubscription(c *gin.Context) {
	var req models.PushSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.SendErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := services.ExtractUserID(c)
	if err != nil {
		models.SendErrorResponse(c, http.StatusUnauthorized, "User ID not found in token")
		return
	}

	// Check if subscription already exists for this user and endpoint
	existingSub := &db.PushSubscription{}
	err = mgm.Coll(existingSub).First(
		bson.M{"user_id": userId, "endpoint": req.Endpoint},
		existingSub,
	)

	if err == nil {
		// Update existing subscription
		existingSub.P256dh = req.Keys.P256dh
		existingSub.Auth = req.Keys.Auth
		err = mgm.Coll(existingSub).Update(existingSub)
		if err != nil {
			log.Printf("Error updating push subscription: %v\n", err)
			models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update push subscription")
			return
		}
		models.SendSuccessResponse(c, "Push subscription updated successfully", nil)
		return
	} else if err != mongo.ErrNoDocuments {
		log.Printf("Error checking for existing push subscription: %v\n", err)
		models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to check existing push subscription")
		return
	}

	// Create new subscription
	newSub := db.NewPushSubscription(userId, req.Endpoint, req.Keys.P256dh, req.Keys.Auth)
	err = mgm.Coll(newSub).Create(newSub)
	if err != nil {
		log.Printf("Error creating push subscription: %v\n", err)
		models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to register push subscription")
		return
	}

	models.SendSuccessResponse(c, "Push subscription registered successfully", nil)
}

// UpdatePushSubscription godoc
// @Summary      Update Push Subscription
// @Description  Update an existing push subscription for the current user
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Subscription ID"
// @Param        req  body      models.PushSubscriptionRequest true "Update Push Subscription Request"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Failure      500  {object}  models.Response
// @Router       /user/push-subscription/{id} [put]
// @Security     ApiKeyAuth
func UpdatePushSubscription(c *gin.Context) {
	var request models.PushSubscriptionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		models.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	userId, exists := c.Get("userId")
	if !exists {
		models.SendErrorResponse(c, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	subscriptionId := c.Param("id")
	if subscriptionId == "" {
		models.SendErrorResponse(c, http.StatusBadRequest, "Subscription ID is required")
		return
	}

	userObjID, ok := userId.(primitive.ObjectID)
	if !ok {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID in context")
		return
	}

	err := services.UpdatePushSubscription(userObjID, subscriptionId, request.Endpoint, request.Keys.P256dh, request.Keys.Auth)
	if err != nil {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update push subscription")
		return
	}

	models.SendSuccessResponse(c, "Push subscription updated successfully", nil)
}

// GetPushSubscriptions godoc
// @Summary      Get Push Subscriptions
// @Description  Get all push subscriptions for the current user
// @Tags         user
// @Produce      json
// @Success      200  {object}  models.Response
// @Failure      500  {object}  models.Response
// @Router       /user/push-subscriptions [get]
// @Security     ApiKeyAuth
func GetPushSubscriptions(c *gin.Context) {
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

	subscriptions, err := services.GetUserPushSubscriptions(userObjID)
	if err != nil {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get push subscriptions")
		return
	}

	models.SendSuccessResponse(c, "Push subscriptions retrieved successfully", map[string]any{
		"subscriptions": subscriptions,
		"total_count":   len(subscriptions),
	})
}

// DeregisterPushSubscription godoc
// @Summary      Deregister Push Subscription
// @Description  Remove a specific push subscription for the current user
// @Tags         user
// @Param        id   path      string  true  "Subscription ID"
// @Produce      json
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Failure      500  {object}  models.Response
// @Router       /user/push-subscription/{id} [delete]
// @Security     ApiKeyAuth
func DeregisterPushSubscription(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		models.SendErrorResponse(c, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	subscriptionId := c.Param("id")
	if subscriptionId == "" {
		models.SendErrorResponse(c, http.StatusBadRequest, "Subscription ID is required")
		return
	}

	userObjID, ok := userId.(primitive.ObjectID)
	if !ok {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID in context")
		return
	}

	err := services.DeregisterPushSubscription(userObjID, subscriptionId)
	if err != nil {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to deregister push subscription")
		return
	}

	models.SendSuccessResponse(c, "Push subscription removed successfully", nil)
}

// DeregisterAllPushSubscriptions godoc
// @Summary      Deregister All Push Subscriptions
// @Description  Remove all push subscriptions for the current user
// @Tags         user
// @Produce      json
// @Success      200  {object}  models.Response
// @Failure      500  {object}  models.Response
// @Router       /user/push-subscriptions [delete]
// @Security     ApiKeyAuth
func DeregisterAllPushSubscriptions(c *gin.Context) {
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

	count, err := services.DeregisterAllPushSubscriptions(userObjID)
	if err != nil {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to deregister push subscriptions")
		return
	}

	models.SendSuccessResponse(c, "All push subscriptions removed successfully", map[string]any{
		"removed_count": count,
	})
}

// TestPushNotification godoc
// @Summary      Test Push Notification
// @Description  Send a test notification to a specific subscription
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Subscription ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Failure      500  {object}  models.Response
// @Router       /user/push-subscription/{id}/test [post]
// @Security     ApiKeyAuth
func TestPushNotification(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		models.SendErrorResponse(c, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	subscriptionId := c.Param("id")
	if subscriptionId == "" {
		models.SendErrorResponse(c, http.StatusBadRequest, "Subscription ID is required")
		return
	}

	userObjID, ok := userId.(primitive.ObjectID)
	if !ok {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID in context")
		return
	}

	err := services.SendTestPushNotification(userObjID, subscriptionId)
	if err != nil {
		models.SendErrorResponse(c, http.StatusInternalServerError, "Failed to send test notification")
		return
	}

	models.SendSuccessResponse(c, "Test notification sent successfully", nil)
}