package main

import (
	"./gophers/controllers"
	_ "./gophers/helpers/mimeTypes"
	"./gophers/plate"
	"flag"
	"log"
	"net/http"
)

var (
	listenAddr = flag.String("http", ":8080", "http listen address")

	CorsHandler = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		return
	}
)

const (
	port = "80"
)

func main() {
	flag.Parse()
	server := plate.NewServer("doughboy")

	server.AddFilter(CorsHandler)

	server.Get("/", controllers.Index)

	server.Static("/", "static")

	http.Handle("/", server)

	log.Println("Server running on port " + *listenAddr)

	log.Fatal(http.ListenAndServe(*listenAddr, nil))

}
