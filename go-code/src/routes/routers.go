package routers

import (
	"os"
	"pingserver/controllers"
	firebase "pingserver/firebase_client"
	"pingserver/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitServer(prod *bool, auth *bool) (r *gin.Engine) {
	if *prod {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	router.Static("/home", "./public")

	apiV1 := router.Group("/api/v1")

	if *auth {
		apiV1.Use(firebase.EnsureLoggedIn())
	} else {
		apiV1.Use(func(c *gin.Context) {
			c.Set("uid", os.Getenv("FIREBASE_UID_DEV"))
		})
	}

	users := apiV1.Group("/users")
	{
		users.GET("/:uid", controllers.GetUserBasic)
		users.GET("/:uid/location", controllers.GetUserLocation)
		users.POST("/", controllers.CreateNewUser)
		users.PUT("/", controllers.UpdateUserInfo)
		users.PUT("/location", controllers.SetUserLocation)
		users.PUT("/notification", controllers.SetNotifToken)
	}

	links := apiV1.Group("/links")
	{
		links.GET("/", controllers.GetAllLinks)
		links.GET("/tosocials/:id", controllers.GetToSocials)
		links.GET("/fromsocials/:id", controllers.GetFromSocials)
		links.GET("/location", controllers.GetLastCheckedInLocations)
		links.PATCH("/tosocials/:id", controllers.UpdatePermissions)
	}

	requests := apiV1.Group("/requests")
	{
		requests.POST("/", controllers.SendRequest)
		requests.DELETE("/:rid/decline", controllers.DeclineRequest)
		requests.PATCH("/:rid", controllers.AcceptRequest)
		requests.DELETE("/:rid/delete", controllers.DeleteRequest)
		requests.GET("/pending", controllers.GetOpenReceivedRequests)
		requests.GET("/sent", controllers.GetOpenSentRequests)
	}

	geoPing := apiV1.Group("/geoping")
	{
		geoPing.POST("/:id", controllers.ShareGeoPing)
		geoPing.POST("/", controllers.CreateGeoPing)
		geoPing.DELETE("/:id", controllers.DeleteGeoPing)
	}

	events := apiV1.Group("/events")
	{
		events.DELETE("/:id", controllers.DeleteEvent)
		events.GET("/:id/attendees", controllers.GetAttendees)
		events.POST("/:id", controllers.HandleAttendance)
		events.GET("/:id", controllers.GetEventDetails)
		events.GET("/", controllers.GetUserRelevantEvents)
		events.PUT("/:id", controllers.UpdateEvent)
		events.POST("/:id/invites", controllers.ShareEvent)
		events.POST("/", controllers.CreateEvent)
		events.PATCH(":id/end", controllers.EndEvent)
		events.GET(":id/invites", controllers.GetPrivateEventShares)
	}

	markers := apiV1.Group("/markers")
	{
		markers.GET("/links", controllers.GetLinkMarkers)
		markers.GET("/geopings", controllers.GetGeoPings)
		markers.GET("/events", controllers.GetEvents)
	}

	models.InitCustomEventValidators()

	return router
}
