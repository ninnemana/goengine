package routes

import (
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

func HandleRoute(w http.ResponseWriter, r *http.Request) {
	rt := &RouteTable{}
	rt = start()
	rt.Match(w, r)
}

/*func Home(w http.ResponseWriter, r *http.Request) {
	helpers.DisplayTemplate("index", "templates/index.html", w, make(map[string]interface{}))
}*/
