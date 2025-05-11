package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hejmsdz/goslides/app"
	"github.com/hejmsdz/goslides/database"
	"github.com/hejmsdz/goslides/di"
)

func startPeriodicTasks(container *di.Container) {
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for range ticker.C {
			cleanedUpSessions := container.Live.CleanUp()
			log.Printf("Cleaned up %d idle sessions", cleanedUpSessions)
		}
	}()
}

func main() {
	db := database.InitializeDB(os.Getenv("DATABASE"))
	redis := database.InitializeRedis(os.Getenv("REDIS"))
	container := di.NewContainer(db, redis)
	startPeriodicTasks(container)

	app := app.NewApp(container)
	srv := &http.Server{
		Addr:    getAddr(),
		Handler: app.Handler(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getAddr() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	return fmt.Sprintf(":%s", port)
}
