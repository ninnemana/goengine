package controllers

import (
	//	"fmt"
	"../helpers/plate"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {

	tmpl := plate.NewTemplate(w)

	tmpl.Layout = "layout.html"
	tmpl.Template = "templates/index.html"

	tmpl.DisplayTemplate()
}
