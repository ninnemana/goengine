package routes

import(
	"gophers/helpers"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request){
	helpers.DisplayTemplate("index", "templates/index.html", w, make(map[string]interface{}))
}