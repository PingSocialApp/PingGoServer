package main

import (
	"context"
	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"strings"

	//  "net/http"
	//  "path/filepath"
)

var app *firebase.App

func setupFirebase (){
	opt := option.WithCredentialsFile("../circles-4d081-firebase-adminsdk-rtjsi-51616d71b7.json")
	fbapp, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic(err.Error())
	}else{
		app = fbapp
	}
}

func ensureLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken := c.Request.Header.Get("Authorization")
		authToken = strings.Replace(authToken, "Bearer ", "", 1)

		// TODO Remove later

		//client, err := app.Auth(context.Background())
		//if err != nil {
		//	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		//		"status":  http.StatusInternalServerError,
		//		"message": http.StatusText(http.StatusInternalServerError),
		//	})
		//	return
		//}


		//_, err = client.VerifyIDToken(context.Background(), authToken)
		//if err != nil {
		//	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		//		"status":  http.StatusUnauthorized,
		//		"message": http.StatusText(http.StatusUnauthorized),
		//	})
		//	return
		//}

		c.Next()
	}
}