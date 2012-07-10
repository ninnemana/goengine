package routes

import (
	"gophers/helpers"
	"net/http"
)

func start() *RouteTable {
	rt := &RouteTable{}

	//      var rt = &Route {
	//          Name: "default",
	//          Pattern: "/{controller}/{action}/{id}",
	//			Method: "GET",
	//          Default: map[string]string { "controller": "home", "action": "index", "id": "0", },
	//          Constraint: map[string]string { "id": "\\d+" }
	//      }

	rt.AddRoute(&Route{
		Name:       "default",
		Pattern:    "/{controller}/{action}/{id}",
		Method:     "GET",
		Default:    map[string]string{"controller": "home", "action": "index", "id": "0"},
		Constraint: map[string]string{"id": "\\d+"},
	})

	return rt
}

type WebContext struct {
	w        http.ResponseWriter
	r        *http.Request
	rd       *RouteData
	vb       map[string]interface{}
	layout   string
	template string
}

func HandleRoute(w http.ResponseWriter, r *http.Request) {
	rt := start()
	rt.Match(w, r)
	//ctx := &WebContext{w: w, r: r, rd: rd}
	helpers.DisplayTemplate()
}

/*func Home(w http.ResponseWriter, r *http.Request) {
	helpers.DisplayTemplate("index", "templates/index.html", w, make(map[string]interface{}))
}*/
