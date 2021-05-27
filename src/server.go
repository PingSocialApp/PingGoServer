package main

import (
	"github.com/gin-gonic/gin"
	dbclient "pingserver/db_client"
	"pingserver/handlers"
)

var router *gin.Engine

func main(){
	//initNeo4j()
	//defer dbclient.CloseDriver()

	//firebase.SetupFirebase()

	router = gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	router.Static("/home","./public")

	users := router.Group("/users")
	{
		users.GET("/:uid", handlers.GetUserBasic)
		users.POST("/:uid", handlers.CreateNewUser)
		users.PUT("/:uid", handlers.UpdateUserInfo)
		users.PUT("/:uid/location", handlers.SetUserLocation)
		users.GET("/:uid/location", handlers.GetUserLocation)
		users.PUT("/:uid/notification", handlers.SetNotifToken)
	}

	links := router.Group("/links")
	{
		links.GET("/", handlers.GetAllLinks)
		links.GET("/:id/tosocials", handlers.GetToSocials)
		links.GET("/:id/fromsocials", handlers.GetFromSocials)
		links.GET("/:id/location", handlers.GetLastCheckedInLocations)
		links.PATCH("/:id/permissions", handlers.UpdatePermissions)
	}

	requests := router.Group("/requests")
	{
		requests.POST("/", handlers.SendRequest)
		requests.DELETE("/:rid/decline", handlers.DeclineRequest)
		requests.PATCH("/:rid", handlers.AcceptRequest)
		requests.DELETE("/:rid/delete", handlers.DeleteRequest)
		requests.GET("/mypending", handlers.GetOpenReceivedRequests)
		requests.GET("/mysent", handlers.GetOpenSentRequests)
		requests.GET("/mypending/num", handlers.GetNumOpenRequests)
	}

	geoPing := router.Group("/geoping")
	{
		geoPing.POST("/:id", handlers.ShareGeoPing)
		geoPing.POST("/", handlers.CreateGeoPing)
		geoPing.DELETE("/:id", handlers.DeleteGeoPing)
	}

	events := router.Group("/events")
	{
		events.DELETE("/:id", handlers.DeleteEvent)
		events.GET("/:id/attendees", handlers.GetAttendees)
		events.POST("/:id", handlers.HandleAttendance)
		events.GET("/:id/details", handlers.GetEventDetails)
		events.GET("/:id/inEventDetails")
		events.GET("/", handlers.GetUserCreatedEvents)
		events.PUT("/:id", handlers.UpdateEvent)
		events.POST("/:id/invites", handlers.ShareEvent)
		events.POST("/", handlers.CreateEvent)
		events.PUT(":id/end", handlers.EndEvent)
		events.GET(":id/invites", handlers.GetPrivateEventShares)
	}

	markers := router.Group("/markers")
	{
		markers.GET("/:uid/links", handlers.GetLinkMarkers)
		markers.GET("/:uid/relevant", handlers.GetRelevantMarkers)
	}

	err := router.Run()
	if err != nil {
		panic(err.Error())
	}

	go handlers.EventCleaner()
}

func initNeo4j(){
	// dbclient.CreateDriver(os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"))
	dbclient.CreateDriver("bolt://localhost:7687", "neo4j", "pingdev")
}