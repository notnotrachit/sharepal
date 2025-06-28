package services

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var Notification *NotificationService

type NotificationService struct {
	app *firebase.App
}

func InitFCM() {
	if Config.FirebaseCredentialsJSON == "" {
		log.Println("Firebase credentials JSON is not set. FCM notifications will be disabled.")
		return
	}

	opt := option.WithCredentialsJSON([]byte(Config.FirebaseCredentialsJSON))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v\n", err)
	}
	Notification = &NotificationService{
		app: app,
	}
}

func (s *NotificationService) SendNotification(token string, title string, body string) error {
	client, err := s.app.Messaging(context.Background())
	if err != nil {
		return err
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Token: token,
	}

	response, err := client.Send(context.Background(), message)
	if err != nil {
		log.Printf("Error sending FCM message: %v\n", err)
		return err
	}
	log.Printf("Successfully sent FCM message: %s\n", response)
	return nil
}
