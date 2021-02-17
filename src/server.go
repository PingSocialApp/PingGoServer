package main

import (
	"github.com/gin-gonic/gin"
	dbclient "pingserver/src/db_client"
	"pingserver/src/handlers"
)

var router *gin.Engine

func main(){
	router := gin.New()
	//dbclient.CreateDriver(os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"))
	dbclient.CreateDriver("bolt://localhost:7687", "neo4j", "pingdev")
	defer dbclient.CloseDriver()
	//setupFirebase()

	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	users := router.Group("/users")
	{
		users.GET("/:uid", ensureLoggedIn(), handlers.GetUserBasic)
		users.GET("/:uid/socials", ensureLoggedIn(), handlers.GetUserSocials)
		users.PATCH("/:uid/checkout", ensureLoggedIn(), handlers.CheckoutUser)
		users.POST("/:uid", handlers.CreateNewUser)
		users.PUT("/:uid", ensureLoggedIn(), handlers.UpdateUserInfo)
		users.PUT("/:uid/notification", ensureLoggedIn(), handlers.SetNotifToken)
	}

	pings := router.Group("/pings", ensureLoggedIn())
	{
		pings.GET("/:uid/num", handlers.GetNumPings)
		pings.GET("/:uid", handlers.GetPings)
		pings.DELETE("/:id", handlers.DeletePing)
		pings.POST("/", handlers.SendPing)
		pings.PUT("/:id", handlers.ReplyPing)
	}

	//links := router.Group("/links", ensureLoggedIn())
	//{
	//	links.DELETE("/:id", handlers.DeleteRequest)
	//	links.GET("/:id/socials", handlers.GetLinkSocials)
	//	links.GET("/:id/permissions", handlers.GetLinkPermissions)
	//	//links.GET("/:uid/all", handlers.GetAllLinks)
	//	//links.GET("/:uid/num", handlers.GetNumPendingLinks)
	//	links.GET("/:uid", handlers.GetSentLinks)
	//	links.GET("/:uid/location", handlers.GetLocationPermittedLinks)
	//	links.PATCH("/:id", handlers.AcceptRequest)
	//	links.PATCH("/:id/permissions", handlers.UpdatePermissions)
	//	links.PUT("/sendRequest/:uid", handlers.SendRequest)
	//}

	//geoPing := router.Group("/geoping", ensureLoggedIn())
	//{
	//	geoPing.GET("geoping/:uid")
	//	geoPing.POST("geoping")
	//}
	//
	//events := router.Group("/events", ensureLoggedIn())
	//{
	//	events.DELETE("events/:id")
	//	events.GET("events/:id/attendees")
	//	events.GET("events/:id/details")
	//	events.GET("events/:id/inPartyDetails")
	//	events.GET("events/:uid")
	//	events.PUT("events/:id")
	//}

	err := router.Run()
	if err != nil {
		panic(err.Error())
	}
}