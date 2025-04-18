package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hejmsdz/goslides/app"
	"github.com/hejmsdz/goslides/database"
	"github.com/hejmsdz/goslides/di"
)

func main() {
	db := database.InitializeDB(os.Getenv("DATABASE"))
	container := di.NewContainer(db)

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
