package main

import (
	"gophers/controllers"
	"gophers/helpers"
	"gophers/routes"
	"net/http"
)

func init() {
	mux := routes.New()

	mux.Get("/", controllers.Index)

	session_key := "your key here"
	http.Handle("/", helpers.NewSessionHandler(mux, key, nil))
}
