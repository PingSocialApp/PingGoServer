package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pingserver/controllers"
	dbclient "pingserver/db_client"
	firebase "pingserver/firebase_client"
	routers "pingserver/routes"

	"github.com/joho/godotenv"
	"github.com/robfig/cron"
)

var c *cron.Cron

func main() {
	localDev := flag.Bool("localDev", true, "local development")
	cloudDB := flag.Bool("cloud", false, "cloud database instance")
	prod := flag.Bool("prod", false, "production mode")
	auth := flag.Bool("auth", true, "route through firebase auth")

	flag.Parse()

	if *localDev {
		err := godotenv.Load()
		if err != nil {
			panic("Error loading .env file")
		}
	}

	if *prod {
		*cloudDB = true
		*localDev = false
		log.Println("Production Instance Setup")
	} else if !(*cloudDB) {
		log.Println("Local Instance Setup")
	} else {
		log.Println("Cloud Instance Setup")
	}

	dbclient.InitNeo4j(cloudDB)

	controllers.Init()

	firebase.SetupFirebase()

	setupCron()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	port, exists := os.LookupEnv("PORT")
	if !exists || port == "" {
		port = ":80"
	}

	log.Println(port)

	srv := &http.Server{
		Addr:    port,
		Handler: routers.InitServer(prod, auth),
	}

	go func() {
		<-quit
		log.Println("receive interrupt signal")
		c.Stop()
		dbclient.CloseDriver()
		if err := srv.Close(); err != nil {
			log.Fatal("Server Close:", err)
		}
		os.Exit(0)
	}()

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err.Error())
	}

}

func setupCron() {
	log.Println("Starting CRON functions")

	c = cron.New()

	err := c.AddFunc("@every 1m", controllers.ExpireEvent)
	if err != nil {
		log.Fatal(err.Error())
	}

	c.Start()
}
