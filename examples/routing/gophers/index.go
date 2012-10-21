package main

import (
	"gophers/controllers"
	"gophers/plate"
	"net/http"
)

func init() {
	server := plate.NewServer("doughboy")

	server.Get("/home", controllers.World)

	// This will get rendered if none of the about routes pass
	server.Get("/", controllers.Index)

	session_key := "your key here"
	http.Handle("/", server.NewSessionHandler(session_key, nil))
}
