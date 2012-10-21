package controllers

import (
	//	"fmt"
	"fmt"
	"gophers/plate"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintf(w, "Hello world!")
	plate.Serve404(w, "")
}

func World(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello world!")
}
