# GoEngine Boilerplate

GoEngine Boilerplate is combination of serveral different repositories, some made for Go some not.

The project is built using routes.go for Sinatra/ExpressJS style routing on the Google App Engine PaaS. HTML Structure and design using h5bp and Twitter Bootstrap. Javascript templating is done with mustache.js, but has not been baked into requireJS. The javascript component incorporates RequireJS for dependency management.

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

# [HTML5 Boilerplate](http://html5boilerplate.com)

HTML5 Boilerplate is a professional front-end template that helps you build fast, robust, adaptable, and future-proof websites. Spend more time developing and less time reinventing the wheel.

This project is the product of many years of iterative development and combined community knowledge. It does not impose a specific development philosophy or framework, so you're free to architect your code in the way that you want.


## Quick start

Clone the git repo - `git clone git://github.com/h5bp/html5-boilerplate.git` - or [download it](https://github.com/h5bp/html5-boilerplate/zipball/master)


## Features

* HTML5 ready. Use the new elements with confidence.
* Cross-browser compatible (Chrome, Opera, Safari, Firefox 3.6+, IE6+).
* Designed with progressive enhancement in mind.
* CSS normalizations and common bug fixes.
* IE-specific classes for easier cross-browser control.
* A default print stylesheet, performance optimized.
* Mobile browser optimizations.
* Protection against any stray `console.log` causing JavaScript errors in IE6/7.
* The latest jQuery via CDN, with a local fallback.
* A custom Modernizr build for feature detection.
* An optimized Google Analytics snippet.
* Apache server caching, compression, and other configuration defaults for Grade-A performance.
* Cross-domain Ajax and Flash.
* "Delete-key friendly." Easy to strip out parts you don't need.
* Extensive inline and accompanying documentation.


## Contributing

Anyone and everyone is welcome to [contribute](https://github.com/h5bp/html5-boilerplate/wiki/contribute). Hundreds of developers have helped make the HTML5 Boilerplate what it is today.


## Project information

* Source: http://github.com/h5bp/html5-boilerplate
* Web: http://html5boilerplate.com
* Docs: http://html5boilerplate.com/docs
* Twitter: http://twitter.com/h5bp


## License

### Major components:

* jQuery: MIT/GPL license
* Modernizr: MIT/BSD license
* Normalize.css: Public Domain

### Everything else:

The Unlicense (aka: public domain)

[Twitter Bootstrap](http://twitter.github.com/bootstrap) [![Build Status](https://secure.travis-ci.org/twitter/bootstrap.png)](http://travis-ci.org/twitter/bootstrap)
=================

Bootstrap provides simple and flexible HTML, CSS, and Javascript for popular user interface components and interactions. In other words, it's a front-end toolkit for faster, more beautiful web development. It's created and maintained by [Mark Otto](http://twitter.com/mdo) and [Jacob Thornton](http://twitter.com/fat) at Twitter.

To get started, checkout http://twitter.github.com/bootstrap!



Quick start
-----------

Clone the repo, `git clone git://github.com/twitter/bootstrap.git`, or [download the latest release](https://github.com/twitter/bootstrap/zipball/master).



Versioning
----------

For transparency and insight into our release cycle, and for striving to maintain backward compatibility, Bootstrap will be maintained under the Semantic Versioning guidelines as much as possible.

Releases will be numbered with the follow format:

`<major>.<minor>.<patch>`

And constructed with the following guidelines:

* Breaking backward compatibility bumps the major (and resets the minor and patch)
* New additions without breaking backward compatibility bumps the minor (and resets the patch)
* Bug fixes and misc changes bumps the patch

For more information on SemVer, please visit http://semver.org/.



Bug tracker
-----------

Have a bug? Please create an issue here on GitHub! Also, when filing please make sure you're familiar with [necolas's guidelines](https://github.com/necolas/issue-guidelines). thanks! <3

https://github.com/twitter/bootstrap/issues



Twitter account
---------------

Keep up to date on announcements and more by following Bootstrap on Twitter, [@TwBootstrap](http://twitter.com/TwBootstrap).



Blog
----

Read more detailed announcements, discussions, and more on [The Official Twitter Bootstrap Blog](http://blog.getbootstrap.com).



Mailing list
------------

Have a question? Ask on our mailing list!

twitter-bootstrap@googlegroups.com

http://groups.google.com/group/twitter-bootstrap



IRC
---

Server: irc.freenode.net

Channel: ##twitter-bootstrap (the double ## is not a typo)



Developers
----------

We have included a makefile with convenience methods for working with the Bootstrap library.

+ **dependencies**
Our makefile depends on you having recess, uglify.js, and jshint installed. To install, just run the following command in npm:

```
$ npm install recess uglify-js jshint -g
```

+ **build** - `make`
Runs the recess compiler to rebuild the `/less` files and compiles the docs pages. Requires recess and uglify-js. <a href="http://twitter.github.com/bootstrap/less.html#compiling">Read more in our docs &raquo;</a>

+ **test** - `make test`
Runs jshint and qunit tests headlessly in phantom js (used for ci). Depends on having phatomjs installed.

+ **watch** - `make watch`
This is a convenience method for watching just Less files and automatically building them whenever you save. Requires the Watchr gem.


Contributing
------------

Please make all pull requests against wip-* branches. Also, if your unit test contains javascript patches or features - you must include relevant unit tests. Thanks!


Authors
-------

**Mark Otto**

+ http://twitter.com/mdo
+ http://github.com/markdotto

**Jacob Thornton**

+ http://twitter.com/fat
+ http://github.com/fat



Copyright and license
---------------------

Copyright 2012 Twitter, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this work except in compliance with the License.
You may obtain a copy of the License in the LICENSE file, or at:

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

# routes.go
a simple http routing API for the Go programming language

    go get github.com/bradrydzewski/routes

this project combines the best of `web.go` and `pat.go`. It uses `pat.go`'s named url parameters (ie `:param`) and `web.go`'s regular expression groups for url matching and parameter extraction, which provides significant performance improvements.

for more information see:
http://gopkgdoc.appspot.com/pkg/github.com/bradrydzewski/routes

## Getting Started

    package main

    import (
        "fmt"
        "github.com/bradrydzewski/routes"
        "net/http"
    )

    func Whoami(w http.ResponseWriter, r *http.Request) {
        params := r.URL.Query()
        lastName := params.Get(":last")
        firstName := params.Get(":first")
        fmt.Fprintf(w, "you are %s %s", firstName, lastName)
    }

    func main() {
        mux := routes.New()
        mux.Get("/:last/:first", Whoami)

        http.Handle("/", mux)
        http.ListenAndServe(":8088", nil)
    }

### Route Examples
You can create routes for all http methods:

    mux.Get("/:param", handler)
    mux.Put("/:param", handler)
    mux.Post("/:param", handler)
    mux.Patch("/:param", handler)
    mux.Del("/:param", handler)

You can specify custom regular expressions for routes:

    mux.Get("/files/:param(.+)", handler)

You can also create routes for static files:

    pwd, _ := os.Getwd()
    mux.Static("/static", pwd)

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