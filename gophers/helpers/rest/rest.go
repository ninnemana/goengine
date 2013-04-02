package rest

import (
	"io/ioutil"
	"net/http"
)

func Get(url string, r *http.Request) (buf []byte, err error) {

	res, err := http.Get(url)
	if err != nil {
		return
	}
	buf, err = ioutil.ReadAll(res.Body)
	res.Body.Close()

	return
}
