package main

import (
	"fmt"
	"os"

	"github.com/hejmsdz/goslides/di"
)

func main() {
	db := InitializeDB(os.Getenv("DATABASE"))
	container := di.NewContainer(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	Server{
		container: container,
		addr:      fmt.Sprintf(":%s", port),
	}.Run()
}

/*
	server := NewServer(container)
	server.Run(fmt.Sprintf(":%s", port))
*/
