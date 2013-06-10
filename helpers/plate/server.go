package plate

import (
	"../mimetypes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	//"webTime"
)

const (
	CONNECT = "CONNECT"
	DELETE  = "DELETE"
	GET     = "GET"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
	PATCH   = "PATCH"
	POST    = "POST"
	PUT     = "PUT"
	TRACE   = "TRACE"

	// log format, modeled after http://wiki.nginx.org/HttpLogModule
	LOG = `%s - - [%s] "%s %s %s" %d %d "%s" "%s"`

	blockSize = 16 // we want 16 byte blocks, for AES-128
)

var (
	// if you don't need multiple independent seshcookie
	// instances, you can use this RequestSessions instance to
	// manage & access your sessions.  Simply use it as the final
	// parameter in your call to seshcookie.NewSessionHandler, and
	// whenever you want to access the current session from an
	// embedded http.Handler you can simply call:
	//
	//     seshcookie.Session.Get(req)
	Session = &RequestSessions{HttpOnly: true}

	// Hash validation of the decrypted cookie failed. Most likely
	// the session was encoded with a different cookie than we're
	// using to decode it, but its possible the client (or someone
	// else) tried to modify the session.
	HashError = errors.New("Hash validation failed")

	// The cookie is too short, so we must exit decoding early.
	LenError = errors.New("Bad cookie length")

	// Let's set a global server just to be safe
	mainServer = NewServer()
)

type Server struct {
	Routes         []*Route
	Logging        bool
	Logger         *log.Logger
	SessionHandler *SessionHandler
	Config         *ServerConfig
	Filters        []http.HandlerFunc
	StatusService  *StatusService
}

//responseWriter is a wrapper for the http.ResponseWriter
// to track if response was written to. It also allows us
// to automatically set certain headers, such as Content-Type,
// Access-Control-Allow-Origin, etc.
type responseWriter struct {
	writer  http.ResponseWriter // Writer
	started bool
	size    int
	status  int
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func NewServer(session_key ...string) *Server {
	f, err := os.Create("server.log")
	if err != nil {
		log.New(os.Stdout, err.Error(), log.Ldate|log.Ltime)
	}

	server := &Server{
		Logging: true,
		Logger:  log.New(f, "", log.Ldate|log.Ltime),
	}
	server.SetLogger(server.Logger)
	if len(session_key) != 0 && len(session_key[0]) != 0 {
		server.NewSessionHandler(session_key[0], nil)
	}

	server.StatusService = NewStatusService()
	return server
}

// Adds a new Route to the Handler
func (this *Server) AddRoute(method string, pattern string, handler http.HandlerFunc) *Route {
	//split the url into sections
	parts := strings.Split(pattern, "/")

	//find params that start with ":"
	//replace with regular expressions
	j := 0
	params := make(map[int]string)
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			expr := "([^/]+)"
			//a user may choose to override the defult expression
			// similar to expressjs: ‘/user/:id([0-9]+)’
			if index := strings.Index(part, "("); index != -1 {
				expr = part[index:]
				part = part[:index]
			}
			params[j] = part
			parts[i] = expr
			j++
		}
	}

	//recreate the url pattern, with parameters replaced
	//by regular expressions. then compile the regex
	pattern = strings.Join(parts, "/")
	regex, regexErr := regexp.Compile(pattern)
	if regexErr != nil {
		//TODO add error handling here to avoid panic
		panic(regexErr)
		return nil
	}

	//now create the Route
	route := &Route{}
	route.method = method
	route.regex = regex
	route.handler = makeGzipHandler(handler)
	route.params = params
	route.sensitive = false
	route.contenttype = mimetypes.TextHtml

	//and finally append to the list of Routes
	this.Routes = append(this.Routes, route)

	return route
}

// Adds a new Route for GET requests
func (this *Server) Get(pattern string, handler http.HandlerFunc) *Route {
	return this.AddRoute(GET, pattern, handler)
}

// Adds a new Route for PUT requests
func (this *Server) Put(pattern string, handler http.HandlerFunc) *Route {
	return this.AddRoute(PUT, pattern, handler)
}

// Adds a new Route for DELETE requests
func (this *Server) Del(pattern string, handler http.HandlerFunc) *Route {
	return this.AddRoute(DELETE, pattern, handler)
}

// Adds a new Route for PATCH requests
// See http://www.ietf.org/rfc/rfc5789.txt
func (this *Server) Patch(pattern string, handler http.HandlerFunc) *Route {
	return this.AddRoute(PATCH, pattern, handler)
}

// Adds a new Route for POST requests
func (this *Server) Post(pattern string, handler http.HandlerFunc) *Route {
	return this.AddRoute(POST, pattern, handler)
}

// Adds a new Route for Static http requests. Serves
// static files from the specified directory
func (this *Server) Static(pattern string, dir string) *Route {
	//append a regex to the param to match everything
	// that comes after the prefix
	pattern = pattern + "(.+)"
	return this.AddRoute(GET, pattern, func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Clean(r.URL.Path)
		path = filepath.Join(dir, path)
		ext := filepath.Ext(path)

		w.Header().Set("Content-Type", mime.TypeByExtension(ext))
		http.ServeFile(w, r, path)
	})
}

// Add middleware filter globally to server
func (this *Server) AddFilter(filter http.HandlerFunc) {
	this.Filters = append(this.Filters, filter)
}

// FIlterParam adds the middleware filter if the REST URL parameter exists.
func (this *Server) FilterParam(param string, filter http.HandlerFunc) {
	if !strings.HasPrefix(param, ":") {
		param = ":" + param
	}

	this.AddFilter(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Query().Get(param)
		if len(p) > 0 {
			filter(w, r)
		}
	})
}

// Required by http.Handler interface. This method is invoked by the
// http server and will handle all page routing
func (this *Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	start_time := time.Now()
	requestPath := r.URL.Path

	//wrap the response writer, in our custom interface
	w := &responseWriter{writer: rw}

	//find a matching Route
	for _, route := range this.Routes {

		//if the methods don't match, skip this handler
		//i.e if request.Method is 'PUT' Route.Method must be 'PUT'
		if r.Method != route.method {
			continue
		}

		// check if route is case sensitive or not
		if route.sensitive == false {
			str := route.regex.String()
			reg, err := regexp.Compile(strings.ToLower(str))

			if err == nil {
				route.regex = reg
				requestPath = strings.ToLower(requestPath)
			}
		}

		//check if Route pattern matches url
		if !route.regex.MatchString(requestPath) {
			continue
		}

		//get submatches (params)
		matches := route.regex.FindStringSubmatch(requestPath)

		//double check that the Route matches the URL pattern.
		if len(matches[0]) != len(requestPath) {
			continue
		}

		if len(route.params) > 0 {
			//add url parameters to the query param map
			values := r.URL.Query()
			for i, match := range matches[1:] {
				values.Add(route.params[i], match)
			}

			//reassemble query params and add to RawQuery
			r.URL.RawQuery = url.Values(values).Encode() + "&" + r.URL.RawQuery
			//r.URL.RawQuery = url.Values(values).Encode()
		}

		if !route.unfiltered {
			// execute global middleware filters
			for _, filter := range this.Filters {
				//go func() {
				filter(w, r)
				//}()
				if w.started {
					return
				}
			}
		}

		//execute middleware filters for this route
		for _, filter := range route.filters {
			go func() {
				filter(w, r)
			}()
			if w.started {
				return
			}
		}
		w.Header().Set("Content-Type", route.contenttype)

		//Invoke the request handler
		route.handler(w, r)
		break
	}

	//if no matches to url, throw a not found exception
	if w.started == false {
		http.NotFound(w, r)
	}

	end_time := time.Now()
	dur := end_time.Sub(start_time)
	this.StatusService.Update(w.status, &dur)

	//if logging is turned on
	if this.Logging {
		this.Logger.Printf(LOG, r.RemoteAddr, time.Now().String(), r.Method,
			r.URL.Path, r.Proto, w.status, w.size,
			r.Referer(), r.UserAgent())
	}
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func makeGzipHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		fn(gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
	}
}

// ---------------------------------------------------------------------------------
// Simple wrapper around a ResponseWriter

// Header returns the header map that will be sent by WriteHeader.
func (this *responseWriter) Header() http.Header {
	return this.writer.Header()
}

// Write writes the data to the connection as part of an HTTP reply,
// and sets `started` to true
func (this *responseWriter) Write(p []byte) (int, error) {
	this.size += len(p)
	this.started = true
	return this.writer.Write(p)
}

// WriteHeader sends an HTTP response header with status code,
// and sets `started` to true
func (this *responseWriter) WriteHeader(code int) {
	this.status = code
	this.started = true
	this.writer.WriteHeader(code)
}

// ---------------------------------------------------------------------------------
// Below are helper functions to replace boilerplate
// code that serializes resources and writes to the
// http response.

// ServeJson replies to the request with a JSON
// representation of resource v.
func ServeJson(w http.ResponseWriter, v interface{}) {
	content, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", mimetypes.ApplicationJson)
	w.Write(content)
}

// ReadJson will parses the JSON-encoded data in the http
// Request object and stores the result in the value
// pointed to by v.
func ReadJson(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// ServeXml replies to the request with an XML
// representation of resource v.
func ServeXml(w http.ResponseWriter, v interface{}) {
	content, err := xml.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", mimetypes.TextXml)
	w.Write(content)
}

// ReadXml will parses the XML-encoded data in the http
// Request object and stores the result in the value
// pointed to by v.
func ReadXml(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	return xml.Unmarshal(body, v)
}

// ServeFormatted replies to the request with
// a formatted representation of resource v, in the
// format requested by the client specified in the
// Accept header.
func ServeFormatted(w http.ResponseWriter, r *http.Request, v interface{}) {
	accept := r.Header.Get("Accept")
	switch accept {
	case mimetypes.ApplicationJsonShort:
		ServeJson(w, v)
	case mimetypes.ApplicationXmlShort, mimetypes.TextXmlShort:
		ServeXml(w, v)
	default:
		ServeJson(w, v)
	}

	return
}

/* End Routing
   ------------------------------- */

/* Logging
   ------------------------------ */

func (this *Server) SetLogger(logger *log.Logger) {
	this.Logger = logger
}

func SetLogger(logger *log.Logger) {
	mainServer.Logger = logger
}

/* Some useful stuff
   -------------------------------- */

type ServerConfig struct {
	StaticDir    string
	Addr         string
	Port         int
	CookieSecret string
	RecoverPanic bool
}

func Serve404(w http.ResponseWriter, error string) {
	bag := make(map[string]interface{})

	tmpl, err := template.New("404.html").ParseFiles("templates/404.html")
	if err != nil {
		log.Println(err)
		return
	}

	err = tmpl.Execute(w, bag)
	return
}

type StatusService struct {
	Lock              sync.Mutex
	Start             time.Time
	Pid               int
	ResponseCounts    map[string]int
	TotalResponseTime time.Time
}

func NewStatusService() *StatusService {
	return &StatusService{
		Start:             time.Now(),
		Pid:               os.Getpid(),
		ResponseCounts:    map[string]int{},
		TotalResponseTime: time.Time{},
	}
}

func (self *StatusService) Update(status_code int, response_time *time.Duration) {
	self.Lock.Lock()
	self.ResponseCounts[fmt.Sprintf("%d", status_code)]++
	self.TotalResponseTime = self.TotalResponseTime.Add(*response_time)
	self.Lock.Unlock()
}

type Status struct {
	Pid                    int
	UpTime                 string
	UpTimeSec              float64
	Time                   string
	TimeUnix               int64
	StatusCodeCount        map[string]int
	TotalCount             int
	TotalResponseTime      string
	TotalResponseTimeSec   float64
	AverageResponseTime    string
	AverageResponseTimeSec float64
}

func (self *StatusService) GetStatus(w http.ResponseWriter, r *http.Request) {

	now := time.Now()

	uptime := now.Sub(self.Start)

	total_count := 0
	for _, count := range self.ResponseCounts {
		total_count += count
	}

	TotalResponseTime := self.TotalResponseTime.Sub(time.Time{})

	average_response_time := time.Duration(0)
	if total_count > 0 {
		avg_ns := int64(TotalResponseTime) / int64(total_count)
		average_response_time = time.Duration(avg_ns)
	}

	st := &Status{
		Pid:                    self.Pid,
		UpTime:                 uptime.String(),
		UpTimeSec:              uptime.Seconds(),
		Time:                   now.String(),
		TimeUnix:               now.Unix(),
		StatusCodeCount:        self.ResponseCounts,
		TotalCount:             total_count,
		TotalResponseTime:      TotalResponseTime.String(),
		TotalResponseTimeSec:   TotalResponseTime.Seconds(),
		AverageResponseTime:    average_response_time.String(),
		AverageResponseTimeSec: average_response_time.Seconds(),
	}

	jsonBytes, err := json.Marshal(st)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	// var buf bytes.Buffer
	// buf.Write(jsonBytes)
	_, err = w.Write(jsonBytes)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
