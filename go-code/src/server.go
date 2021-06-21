package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	dbclient "pingserver/db_client"
	firebase "pingserver/firebase_client"
	"pingserver/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// var router *gin.Engine

func main() {
	cloudDB := flag.Bool("cloud", false, "cloud database instance")
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	if !(*cloudDB) {
		fmt.Println("Local Instance Setup")
	} else {
		fmt.Println("Cloud Instance Setup")
	}

	initNeo4j(cloudDB)
	defer dbclient.CloseDriver()

	firebase.SetupFirebase()

	err = initServer().Run()
	if err != nil {
		log.Fatalf(err.Error())
	}

	go handlers.EventCleaner()
}

func initNeo4j(cloudDB *bool) {
	if *cloudDB {
		dbclient.CreateDriver(os.Getenv("CLOUD_DEV_URL"), os.Getenv("CLOUD_DEV_USER"), os.Getenv("CLOUD_DEV_PASS"))
	} else {
		dbclient.CreateDriver(os.Getenv("LOCAL_DEV_URL"), os.Getenv("LOCAL_DEV_USER"), os.Getenv("LOCAL_DEV_PASS"))
	}
}

func initServer() (r *gin.Engine) {
	router := gin.New()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(firebase.EnsureLoggedIn())

	router.Static("/home", "./public")

	users := router.Group("/users")
	{
		users.GET("/:uid", handlers.GetUserBasic)
		users.GET("/:uid/location", handlers.GetUserLocation)
		users.POST("/", handlers.CreateNewUser)
		users.PUT("/", handlers.UpdateUserInfo)
		users.PUT("/location", handlers.SetUserLocation)
		users.PUT("/notification", handlers.SetNotifToken)
	}

	links := router.Group("/links")
	{
		links.GET("/", handlers.GetAllLinks)
		links.GET("/tosocials/:id", handlers.GetToSocials)
		links.GET("/fromsocials/:id", handlers.GetFromSocials)
		links.GET("/location", handlers.GetLastCheckedInLocations)
		links.PATCH("/tosocials/:id", handlers.UpdatePermissions)
	}

	requests := router.Group("/requests")
	{
		requests.POST("/", handlers.SendRequest)
		requests.DELETE("/:rid/decline", handlers.DeclineRequest)
		requests.PATCH("/:rid", handlers.AcceptRequest)
		requests.DELETE("/:rid/delete", handlers.DeleteRequest)
		requests.GET("/pending", handlers.GetOpenReceivedRequests)
		requests.GET("/sent", handlers.GetOpenSentRequests)
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
		events.GET("/", handlers.GetUserCreatedEvents)
		events.PUT("/:id", handlers.UpdateEvent)
		events.POST("/:id/invites", handlers.ShareEvent)
		events.POST("/", handlers.CreateEvent)
		events.PATCH(":id/end", handlers.EndEvent)
		events.GET(":id/invites", handlers.GetPrivateEventShares)
	}

	markers := router.Group("/markers")
	{
		markers.GET("/links", handlers.GetLinkMarkers)
		markers.GET("/geopings", handlers.GetGeoPings)
		markers.GET("/events", handlers.GetEvents)
	}
	return router
}
