package firebase_client

import (
	"context"
	"log"

	"firebase.google.com/go/messaging"
)

type Message struct {
	Title string
	Body  string
	Data  map[string]string
}

type NotifTarget struct {
	Ids    []string
	Tokens []string
}

func SendSingleNotif(registrationToken string, data *Message) {
	ctx := context.Background()

	// See documentation on defining a message payload.
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: data.Title,
			Body:  data.Body,
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
		Data:  data.Data,
		Token: registrationToken,
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	response, err := Messaging.Send(ctx, message)
	if err != nil {
		log.Println(err.Error())
	} else {
		// Response is a message ID string.
		log.Println("Successfully sent message:", response)
	}

}

func SendMultiNotif(registrationTokens []string, data *Message) {
	ctx := context.Background()

	// See documentation on defining a message payload.
	message := &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: data.Title,
			Body:  data.Body,
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
		Data:   data.Data,
		Tokens: registrationTokens,
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	response, err := Messaging.SendMulticast(ctx, message)
	if err != nil {
		log.Println(err.Error())
		return
	} else {
		// Response is a message ID string.
		log.Println("Successfully sent message:", response)
	}
}
