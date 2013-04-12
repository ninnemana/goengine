# GoEngine

GoEngine is combination of serveral different repositories, some made for Go some not.

The project is built using routes.go for Sinatra/ExpressJS style routing on the Google App Engine PaaS. HTML Structure and design using h5bp and Twitter Bootstrap. Javascript templating is done with mustache.js, but has not been baked into requireJS. The javascript component incorporates RequireJS for dependency management.

**IMPORTANT**: If you are using Google AppEngine, use the [_gae_](https://github.com/ninnemana/goengine/tree/gae) branch. If you are **__not__** using Google AppEngine, use the _master_ branch.

Mustache.js
-----------

mustache.js has been converted to use [[ ]] as delimiters so it can play nice with golang's html/template package.

Issues
-----------

There is currently (1.7.0) an issue with passing routes with spaces on the App Engine dev_appserver.py. The issue does not seem to exist on the live server. We have found that making a small change to /google/appengine/ext/go/__init__.py will resolve this issue.

Remove from line 513:
```
request_uri = env['PATH_INFO']
```

Replace with:
```
request_uri = env['_AH_ENCODED_SCRIPT_NAME']
```

Contributors
-----------

**Alex Ninneman**

+ http://twitter.com/ninnemana
+ http://github.com/ninnemana

**Jessica Janiuk**

+ http://twitter.com/janiukjf
+ http://github.com/janiukjf

## Getting Started

    package main

    import (
        "gophers/plate"
        "net/http"
    )

    func Whoami(w http.ResponseWriter, r *http.Request) {
        params := r.URL.Query()
        first := params.Get(":first")
        last := params.Get(":last")
        fmt.Fprintf(w, "Hello, %s %s", first, last)
    }

    func init() {
        server := plate.NewServer("doughboy")
        server.Get("/", Whoami)
        session_key := "your_key_here"
        http.Handle("/",server.NewSessionHandler(session_key, nil))
    }

### Route Examples
You can create routes for all http methods:

    server.Get("/:param", handler)
    server.Put("/:param", handler)
    server.Post("/:param", handler)
    server.Patch("/:param", handler)
    server.Del("/:param", handler)

You can specify custom regular expressions for routes:

    server.Get("/files/:param(.+)", handler)

You can also create routes for static files:

    pwd, _ := os.Getwd()
    server.Static("/static", pwd)

this will serve any files in `/static`, including files in subdirectories. For example `/static/logo.gif` or `/static/style/main.css`.

## Helper Functions
You can use helper functions for serializing to Json and Xml. I found myself constantly writing code to serialize, set content type, content length, etc. Feel free to use these functions to eliminate redundant code in your app.

Helper function for serving Json, sets content type to `application/json`:

    func handler(w http.ResponseWriter, r *http.Request) {
        mystruct := { ... }
        routes.ServeJson(w, &mystruct)
    }

Helper function for serving Xml, sets content type to `application/xml`:

    func handler(w http.ResponseWriter, r *http.Request) {
        mystruct := { ... }
        routes.ServeXml(w, &mystruct)
    }

Helper function to serve Xml OR Json, depending on the value of the `Accept` header:

    func handler(w http.ResponseWriter, r *http.Request) {
        mystruct := { ... }
        routes.ServeFormatted(w, r, &mystruct)
    }

## Security
You can restrict access to routes by assigning an `AuthHandler` to a route.

Here is an example using a custom `AuthHandler` per route. Image we are doing some type of Basic authentication:

    func authHandler(w http.ResponseWriter, r *http.Request) bool {
        user := r.URL.User.Username()
        password := r.URL.User.Password()
        if user != "xxx" && password != "xxx" {
            // if we wanted, we could do an http.Redirect here
            return false
        }
        return true
    }

    mux.Get("/:param", handler).SecureFunc(authHandler)

If you plan to use the same `AuthHandler` to secure all of your routes, you may want to set the `DefaultAuthHandler`:

    routes.DefaulAuthHandler = authHandler
    mux.Get("/:param", handler).Secure()
    mux.Get("/:param", handler).Secure()

### OAuth2
In the above examples, we implemented our own custom `AuthHandler`. Check out the [auth.go](https://github.com/bradrydzewski/auth.go) API which provides custom AuthHandlers for OAuth2 providers such as Google and Github.

## Logging
Logging is enabled by default, but can be disabled:

    mux.Logging = false

You can also specify your logger:

    mux.Logger = log.New(os.Stdout, "", 0)
