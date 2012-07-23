package gophers

import (
	"net/http"
)

func homeActions(r *http.Request, rd *RouteData) map[string]*Action {

	actionlist := make(map[string]*Action)

	actionlist["index"] = &Action{
		Name: "index",
		Run:  homeIndex(r, rd),
	}

	return actionlist
}

func homeIndex(r *http.Request, rd *RouteData) map[string]interface{} {

	//fmt.Println(r.URL.Parse)

	vb := make(map[string]interface{})
	return vb
}
