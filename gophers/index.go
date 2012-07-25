package main

import (
	"gophers/controllers"
	"gophers/routes"
	"net/http"
)

func init() {
	mux := routes.New()

	mux.Get("/understanding%20towing", controllers.Understand)
	mux.Get("/", controllers.Index)
	http.Handle("/", mux)
}
