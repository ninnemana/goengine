package controllers

import (
	//	"fmt"
	"../helpers/plate"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	server := plate.NewServer()

	tmpl, _ := server.Template(w)

	tmpl.Template = "templates/index.html"

	tmpl.DisplayTemplate()
}