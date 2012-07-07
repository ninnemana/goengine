package site

import(
	"gophers/routes"
	"net/http"
)


func init(){
	http.HandleFunc("/", routes.Home)
}