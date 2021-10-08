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
	"pingserver/queue"
	routers "pingserver/routes"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	dev := flag.Bool("dev", false, "dev-release mode")
	prod := flag.Bool("prod", false, "production-release mode")
	auth := flag.Bool("auth", true, "route through firebase auth")

	flag.Parse()

	if *dev {
		err := godotenv.Load()
		if err != nil {
			panic("Error loading .env file")
		}
	}

	if *prod {
		log.Println("Production-Release Instance Setup")
	} else if *dev {
		log.Println("Dev-Release Instance Setup")
	} else {
		log.Println("Dev Instance Setup")
	}

	dbclient.InitNeo4j()

	controllers.Init()

	firebase.SetupFirebase()

	queue.InitDispatcher()
	initCron()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	port, exists := os.LookupEnv("PORT")
	if !exists || port == "" {
		port = "80"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: routers.InitServer(prod, auth),
	}

	go func() {
		<-quit
		log.Println("Receive Interrupt Signal")
		queue.Dispatcher.Stop()
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

func initCron() {
	_, err := queue.Dispatcher.DispatchCron(controllers.ExpireEvent, "@every 1m")
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = queue.Dispatcher.DispatchCron(controllers.NotifyEventStart, "@every 1m")
	if err != nil {
		log.Fatal(err.Error())
	}
}
