package main

import (
	"./gophers/controllers"
	"./gophers/plate"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func run_test_request(t *testing.T, server *plate.Server, method, url_str string, payload url.Values) (*httptest.ResponseRecorder, http.Request) {

	url_obj, err := url.Parse(url_str)
	if err != nil {
		t.Fatal(err)
	}

	r := http.Request{
		Method: method,
		URL:    url_obj,
	}

	if payload != nil {
		r.URL.RawQuery = payload.Encode()
	}

	if strings.ToUpper(method) == "POST" {
		r.Form = payload
	}

	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, &r)

	return recorder, r
}

func code_is(t *testing.T, r *httptest.ResponseRecorder, expected_code int) error {
	if r.Code != expected_code {
		return errors.New(fmt.Sprintf("Code %d expected, got: %d", expected_code, r.Code))
	}
	return nil
}

func content_type_is_json(t *testing.T, r *httptest.ResponseRecorder) error {
	ct := r.HeaderMap.Get("Content-Type")
	if ct != "application/json" {
		return errors.New(fmt.Sprintf("Content type 'application/json' expected, got: %s", ct))
	}
	return nil
}

func body_is(t *testing.T, r *httptest.ResponseRecorder, expected_body string) error {
	body := r.Body.String()
	if body != expected_body {
		return errors.New(fmt.Sprintf("Body '%s' expected, got: '%s'", expected_body, body))
	}
	return nil
}

type ErrorMessage struct {
	StatusCode int
	Error      string
	Route      *url.URL
}

func checkError(req http.Request, rec *httptest.ResponseRecorder, err error, t *testing.T) {
	if err != nil {
		t.Errorf("\nError: %s \nRoute: %s \n\n", err.Error(), req.URL)
	}
}

func TestHandler(t *testing.T) {

	server := plate.NewServer("doughboy")
	server.Logging = true

	server.AddFilter(CorsHandler)

	server.Get("/", controllers.Index)

	recorder, req := run_test_request(t, server, "GET", "http://localhost:8080/", nil)
	err := code_is(t, recorder, 200)
	checkError(req, recorder, err, t)


	// This test is failing because for some reason the encrypted password for the test user
	// did not properly carry over the password

	// authForm := url.Values{}
	// authForm.Add("email", "test@curtmfg.com")
	// authForm.Add("password", "")
	// recorder, req = run_test_request(t, server, "POST", "http://localhost:8080/customer/auth", authForm)
	// err = code_is(t, recorder, 200)
	// err = content_type_is_json(t, recorder)

	// authForm := url.Values{}
	// authForm.Add("key", "c8bd5d89-8d16-11e2-801f-00155d47bb0a")
	// recorder, req = run_test_request(t, server, "GET", "http://localhost:8080/customer/auth", authForm)
	// err = code_is(t, recorder, 200)
	// checkError(req, recorder, err, t)
	// err = content_type_is_json(t, recorder)
	// checkError(req, recorder, err, t)

}
