package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"pingserver/db_client"
	"pingserver/handlers"
)

var router *gin.Engine

func main(){
	router := gin.Default()
	dbclient.CreateDriver("bolt://localhost:7687", "neo4j", "pingdev")
	defer dbclient.CloseDriver()

	users := router.Group("/users")
	{
		users.GET("/:uid", handlers.GetUserBasic)
		users.GET("/:uid/socials", handlers.GetUserSocials)
		users.PATCH("/:uid/checkout", handlers.CheckoutUser)	
		users.POST("/:uid", handlers.CreateNewUser)
		users.PUT("/:uid", handlers.UpdateUserInfo)
		users.PUT("/:uid/notification", handlers.SetNotifToken)
	}

	pings := router.Group("/pings")
	{
		pings.GET("/:uid/num", handlers.GetNumPings)
		pings.GET("/:uid", handlers.GetPings)
		pings.DELETE("/:id", handlers.DeletePing)
		pings.POST("/", handlers.SendPing)
		pings.PUT("/:id", handlers.ReplyPing)
	}

	// links := router.Group("/links")
	// {
	// 	links.DELETE("/:id", handlers.DeleteRequest)
	// 	links.GET("/:id/socials", handlers.GetLinkSocials)
	// 	links.GET("/:id/permissions", handlers.GetLinkPermissions)
	// 	links.GET("/:uid/all", handlers.GetAllLinks)
	// 	links.GET("/:uid/num", handlers.GetNumPendingLinks)
	// 	links.GET("/:uid", handlers.GetSentLinks)
	// 	links.GET("/:uid/location", handlers.GetLocationPermittedLinks)
	// 	links.PATCH("/:id", handlers.AcceptRequest)
	// 	links.PATCH("/:id/permissions", handlers.UpdatePermissions)
	// 	links.PUT("/sendRequest/:uid", handlers.SendRequest)	
	// }

	// router.GET("geoping/:uid")
	// router.POST("geoping")

	// events := router.Group("/events")
	// {
	// 	events.DELETE("events/:id")
	// 	events.GET("events/:id/attendees")
	// 	events.GET("events/:id/details")
	// 	events.GET("events/:id/inPartyDetails")
	// 	events.GET("events/:uid")
	// 	events.PUT("events/:id")
	// }

	err := router.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}