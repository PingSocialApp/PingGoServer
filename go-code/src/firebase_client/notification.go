package firebase_client

import (
	"context"
	"firebase.google.com/go/messaging"
	"log"
)

type Message struct {
	title string
	body  string
}

func SendSingleNotif(registrationToken string, data Message) {
	ctx := context.Background()

	// See documentation on defining a message payload.
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: data.title,
			Body:  data.body,
		},
		Token: registrationToken,
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	response, err := Messaging.Send(ctx, message)
	if err != nil {
		log.Println(err.Error())
	}
	// Response is a message ID string.
	log.Println("Successfully sent message:", response)
}

func SendMultiNotif(registrationTokens []string, data Message) {
	ctx := context.Background()

	// See documentation on defining a message payload.
	message := &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: data.title,
			Body:  data.body,
		},
		Tokens: registrationTokens,
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	response, err := Messaging.SendMulticast(ctx, message)
	if err != nil {
		log.Println(err.Error())
	}
	// Response is a message ID string.
	log.Println("Successfully sent message:", response)
}
