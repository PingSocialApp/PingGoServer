package firebase_client

import (
	"context"
	b64 "encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"

	"firebase.google.com/go/db"
	"firebase.google.com/go/messaging"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

var FbClient *firebase.App
var Firestore *firestore.Client
var Messaging *messaging.Client
var RTDB *db.Client

func SetupFirebase() {
	sDec, err := b64.URLEncoding.DecodeString(os.Getenv("ADMIN_SDK"))
	if err != nil {
		log.Fatalf(err.Error())
	}

	opt := option.WithCredentialsJSON(sDec)
	config := &firebase.Config{
		DatabaseURL: "https://circles-4d081.firebaseio.com/",
	}
	fbapp, err := firebase.NewApp(context.Background(), config, opt)
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
		RTDB, err = FbClient.Database(context.Background())
		if err != nil {
			log.Fatalf("error getting RTDB client: %v\n", err.Error())
		}
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
