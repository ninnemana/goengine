package main

import (
	"gophers/controllers"
	"gophers/plate"
	"net/http"
)

var AuthHandler = func(rw http.ResponseWriter, req *http.Request) bool {

	// if user is authenticated we return true, if not return false
	// 
	// You can use serveral different methods to store authentication data
	// in GoEngine
	// 
	// Session: This is demonstrated in the session example
	// 
	// Google AppEngine Memcache: Leverage AppEngine's key/value store

	return true
}

func init() {
	server := plate.NewServer("doughboy")

	// Assigning the AuthenticatedHandler for our secure routing
	plate.DefaultAuthHandler = AuthHandler

	// Demonstrate a GET defined route
	server.Get("/world", controllers.World)

	// A POST route
	server.Post("/world", controllers.PostWorld)

	server.Del("/world/vegetables", controllers.DeleteVegetables) // Delete route
	server.Put("/world/fruit", controllers.PutFruit)              // Put route
	server.Patch("/world/meat", controllers.PatchMeat)            // Patch route

	server.Get("/hello/Sensitive", controllers.Sensi).Sensitive() // Case sensitive route

	// A GET route with URL data
	server.Get("/hello/:name", controllers.ByURL)

	// Authentication Examples
	server.Get("/secure", controllers.Secure).Secure()                        // Uses our global AuthHandler func
	server.Get("/secure/special", controllers.Secure).SecureFunc(AuthHandler) // Specified Auth Handler

	// This will get rendered if none of the about routes pass
	server.Get("/", controllers.Index)

	session_key := "your key here"
	http.Handle("/", server.NewSessionHandler(session_key, nil))
}
