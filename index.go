package main

import (
	"./gophers/controllers"
	"./gophers/helpers/globals"
	_ "./gophers/helpers/mimetypes"
	"./gophers/plate"
	"log"
	"net/http"
)

var (
	CorsHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		return
	}
)

const (
	port = "80"
)

func main() {
	globals.SetGlobals()
	server := plate.NewServer("doughboy")

	server.AddFilter(CorsHandler)

	server.Get("/", controllers.Index)

	server.Static("/", *globals.Filepath+"static")

	http.Handle("/", server)

	log.Println("Server running on port " + *globals.ListenAddr)

	log.Fatal(http.ListenAndServe(*globals.ListenAddr, nil))

}
