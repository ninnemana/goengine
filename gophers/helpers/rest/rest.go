package rest

import (
	"appengine"
	"appengine/urlfetch"
	"bytes"
	"net/http"
)

func Get(url string, r *http.Request) (buf bytes.Buffer, err error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	t := &urlfetch.Transport{Context: appengine.NewContext(r)}

	trip, err := t.RoundTrip(req)
	if err != nil {
		return
	}

	defer trip.Body.Close()

	buf.ReadFrom(trip.Body)

	return
}
