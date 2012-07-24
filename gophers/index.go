package main

import (
	"gophers/controllers"
	"net/http"
)

func init() {
	http.HandleFunc("/", gophers.HandleRoute)
}
