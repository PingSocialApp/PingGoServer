package firebase_client

import (
	"context"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"

	// "firebase.google.com/go/db"
	"firebase.google.com/go/messaging"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

var FbClient *firebase.App
var Firestore *firestore.Client
var Messaging *messaging.Client

// var RTDB *db.Client

func SetupFirebase() {
	//TODO Sub Credentials to OS var
	opt := option.WithCredentialsFile("firebase_client/circles-4d081-firebase-adminsdk-rtjsi-6ab3240fd0.json")
	// config := &firebase.Config{
	// 	DatabaseURL: "https://circles-4d801.firebaseio.com",
	// }
	fbapp, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf(err.Error())
	} else {
		FbClient = fbapp
		Firestore, err = FbClient.Firestore(context.Background())
		if err != nil {
			log.Fatalf("error getting Firestore client: %v\n", err.Error())
		}
		Messaging, err = FbClient.Messaging(context.Background())
		if err != nil {
			log.Fatalf("error getting Messaging client: %v\n", err.Error())
		}
		// RTDB, err = FbClient.Database(context.Background())
		// if err != nil {
		// 	log.Fatalf("error getting RTDB client: %v\n", err.Error())
		// }
	}
}

func EnsureLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken := c.Request.Header.Get("Authorization")
		authToken = strings.Replace(authToken, "Bearer ", "", 1)

		client, err := FbClient.Auth(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": http.StatusInternalServerError,
				"data":  http.StatusText(http.StatusInternalServerError),
			})
			return
		}

		userData, err := client.VerifyIDToken(c, authToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": http.StatusUnauthorized,
				"data":  http.StatusText(http.StatusUnauthorized),
			})
			return
		}

		c.Set("uid", userData.UID)

		c.Next()
	}
}
