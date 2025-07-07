package services

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	db "github.com/ebubekiryigit/golang-mongodb-rest-api-starter/models/db"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var Notification *NotificationService

type NotificationService struct {
	VapidPublicKey  string
	VapidPrivateKey string
}

func InitWebPush() {
	if Config.VapidPublicKey == "" || Config.VapidPrivateKey == "" {
		log.Println("VAPID keys are not set. Push notifications will be disabled.")
		return
	}
	Notification = &NotificationService{
		VapidPublicKey:  Config.VapidPublicKey,
		VapidPrivateKey: Config.VapidPrivateKey,
	}
}

func (s *NotificationService) SendNotification(subscription *db.PushSubscription, message []byte) error {
	sub := &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			Auth:   subscription.Auth,
			P256dh: subscription.P256dh,
		},
	}

	resp, err := webpush.SendNotification(message, sub, &webpush.Options{
		VAPIDPublicKey:  s.VapidPublicKey,
		VAPIDPrivateKey: s.VapidPrivateKey,
		TTL:             60,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 410 || resp.StatusCode == 404 {
		// Subscription is no longer valid, remove it from the database
		log.Printf("Subscription for endpoint %s is no longer valid. Deleting.", subscription.Endpoint)
		// TODO: Implement logic to delete the subscription from the database
	}

	log.Printf("Successfully sent push notification to endpoint: %s\n", subscription.Endpoint)
	return nil
}

// UpdatePushSubscription updates an existing push subscription
func UpdatePushSubscription(userID primitive.ObjectID, subscriptionID, endpoint, p256dh, auth string) error {
	subscription := &db.PushSubscription{}
	
	// Find subscription by ID and user ID
	objID, err := primitive.ObjectIDFromHex(subscriptionID)
	if err != nil {
		return errors.New("invalid subscription ID")
	}
	
	err = mgm.Coll(subscription).First(bson.M{
		"_id":     objID,
		"user_id": userID,
	}, subscription)
	if err != nil {
		return errors.New("subscription not found")
	}
	
	// Update subscription fields
	subscription.Endpoint = endpoint
	subscription.P256dh = p256dh
	subscription.Auth = auth
	
	return mgm.Coll(subscription).Update(subscription)
}

// GetUserPushSubscriptions retrieves all push subscriptions for a user
func GetUserPushSubscriptions(userID primitive.ObjectID) ([]*db.PushSubscription, error) {
	var subscriptions []*db.PushSubscription
	
	err := mgm.Coll(&db.PushSubscription{}).SimpleFind(&subscriptions, bson.M{
		"user_id": userID,
	})
	
	return subscriptions, err
}

// DeregisterPushSubscription removes a specific push subscription
func DeregisterPushSubscription(userID primitive.ObjectID, subscriptionID string) error {
	objID, err := primitive.ObjectIDFromHex(subscriptionID)
	if err != nil {
		return errors.New("invalid subscription ID")
	}
	
	subscription := &db.PushSubscription{}
	err = mgm.Coll(subscription).First(bson.M{
		"_id":     objID,
		"user_id": userID,
	}, subscription)
	if err != nil {
		return errors.New("subscription not found")
	}
	
	return mgm.Coll(subscription).Delete(subscription)
}

// DeregisterAllPushSubscriptions removes all push subscriptions for a user
func DeregisterAllPushSubscriptions(userID primitive.ObjectID) (int, error) {
	result, err := mgm.Coll(&db.PushSubscription{}).DeleteMany(mgm.Ctx(), bson.M{
		"user_id": userID,
	})
	if err != nil {
		return 0, err
	}
	
	return int(result.DeletedCount), nil
}

// SendTestPushNotification sends a test notification to a specific subscription
func SendTestPushNotification(userID primitive.ObjectID, subscriptionID string) error {
	objID, err := primitive.ObjectIDFromHex(subscriptionID)
	if err != nil {
		return errors.New("invalid subscription ID")
	}
	
	subscription := &db.PushSubscription{}
	err = mgm.Coll(subscription).First(bson.M{
		"_id":     objID,
		"user_id": userID,
	}, subscription)
	if err != nil {
		return errors.New("subscription not found")
	}
	
	// Create test notification
	notification := map[string]interface{}{
		"title": "Test Notification",
		"body":  "This is a test push notification from SharePal",
		"data": map[string]interface{}{
			"type":      "test",
			"timestamp": time.Now().Unix(),
		},
	}
	
	if Notification == nil {
		return errors.New("notification service not initialized")
	}
	
	return Notification.SendJSONNotification(subscription, notification)
}

func (s *NotificationService) SendJSONNotification(subscription *db.PushSubscription, data interface{}) error {
	message, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.SendNotification(subscription, message)
}
