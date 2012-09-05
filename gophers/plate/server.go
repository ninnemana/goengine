package plate

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"hash"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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

	// commonly used mime types
	applicationJson = "application/json"
	applicationXml  = "applicatoin/xml"
	textXml         = "text/xml"

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
}

type Route struct {
	method    string
	regex     *regexp.Regexp
	params    map[int]string
	handler   http.HandlerFunc
	auth      AuthHandler
	sensitive bool
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

type sessionResponseWriter struct {
	http.ResponseWriter
	h   *SessionHandler
	req *http.Request
	// int32 so we can use the sync/atomic functions on it
	wroteHeader int32
}

type SessionHandler struct {
	http.Handler
	CookieName string // name of the cookie to store our session in
	CookiePath string // resource path the cookie is valid for
	RS         *RequestSessions
	encKey     []byte
	hmacKey    []byte
}

type RequestSessions struct {
	HttpOnly bool // don't allow javascript to access cookie
	Secure   bool // only send session over HTTPS
	lk       sync.Mutex
	m        map[*http.Request]map[string]interface{}
	// stores a hash of the serialized session (the gob) that we
	// received with the start of the request.  Before setting a
	// cookie for the reply, check to see if the session has
	// actually changed.  If it hasn't, then we don't need to send
	// a new cookie.
	hm map[*http.Request][]byte
}

type Template struct {
	Layout   string
	Template string
	Bag      map[string]interface{}
	Writer   http.ResponseWriter
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
	route.handler = handler
	route.params = params
	route.sensitive = false

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

// Secures a route using the default AuthHandler
func (this *Route) Secure() *Route {
	this.auth = DefaultAuthHandler
	return this
}

// SecureFunc a route using a custom AuthHandler function
func (this *Route) SecureFunc(handler AuthHandler) *Route {
	this.auth = handler
	return this
}

func (this *Route) Sensitive() *Route {
	this.sensitive = true
	return this
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
		http.ServeFile(w, r, path)
	})
}

// Required by http.Handler interface. This method is invoked by the
// http server and will handle all page routing
func (this *Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	//wrap the response writer, in our custom interface
	w := &responseWriter{writer: rw}

	//find a matching Route
	for _, route := range this.Routes {

		requestPath := r.URL.Path

		//if the methods don't match, skip this handler
		//i.e if request.Method is 'PUT' Route.Method must be 'PUT'
		if r.Method != route.method {
			continue
		}

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

		//add url parameters to the query param map
		values := r.URL.Query()
		for i, match := range matches[1:] {
			values.Add(route.params[i], match)
		}

		//reassemble query params and add to RawQuery

		r.URL.RawQuery = url.Values(values).Encode()

		//enfore security, if necessary
		if route.auth != nil {
			//autenticate the user
			ok := route.auth(w, r)
			//if the auth handler redirected the user
			//or already wrote a response, we can just exit
			if w.started {
				return
			} else if ok == false {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}
		}

		//Invoke the request handler
		route.handler(w, r)

		break
	}

	//if no matches to url, throw a not found exception
	if w.started == false {
		http.NotFound(w, r)
	}

	//if logging is turned on
	if this.Logging {
		this.Logger.Printf(LOG, r.RemoteAddr, time.Now().String(), r.Method,
			r.URL.Path, r.Proto, w.status, w.size,
			r.Referer(), r.UserAgent())
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
// Authentication helper functions to enable user authentication

type AuthHandler func(http.ResponseWriter, *http.Request) bool

// DefaultAuthHandler will be applied to any route when the Secure() function
// is invoked, as opposed to SecureFunc(), which accepts a custom AuthHandler.
//
// By default, the DefaultAuthHandler will deny all requests. This value
// should be replaced with a custom AuthHandler implementation, as this
// is just a dummy function.
var DefaultAuthHandler = func(w http.ResponseWriter, r *http.Request) bool {
	return false
}

// ---------------------------------------------------------------------------------
// Below are helper functions to replace boilerplate
// code that serializes resources and writes to the
// http response.

// ServeJson replies to the request with a JSON
// representation of resource v.
func ServeJson(w http.ResponseWriter, v interface{}) {
	content, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(content)
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", applicationJson)
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
	w.Write(content)
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
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
	case applicationJson:
		ServeJson(w, v)
	case applicationXml, textXml:
		ServeXml(w, v)
	default:
		ServeJson(w, v)
	}

	return
}

/* End Routing
   ------------------------------- */

/* Session worker
   ------------------------------- */

func (rs *RequestSessions) Get(req *http.Request) map[string]interface{} {
	rs.lk.Lock()
	defer rs.lk.Unlock()

	if rs.m == nil {
		log.Printf(LOG, "seshcookie: warning! trying to get session "+
			"data for unknown request. Perhaps your handler "+
			"isn't wrapped by a SessionHandler?")
		return nil
	}

	return rs.m[req]
}

func (rs *RequestSessions) getHash(req *http.Request) []byte {
	rs.lk.Lock()
	defer rs.lk.Unlock()

	if rs.hm == nil {
		return nil
	}

	return rs.hm[req]
}

func (rs *RequestSessions) Set(req *http.Request, val map[string]interface{}, gobHash []byte) {
	rs.lk.Lock()
	defer rs.lk.Unlock()

	if rs.m == nil {
		rs.m = map[*http.Request]map[string]interface{}{}
		rs.hm = map[*http.Request][]byte{}
	}

	rs.m[req] = val
	rs.hm[req] = gobHash
}

func (rs *RequestSessions) Clear(req *http.Request) {
	rs.lk.Lock()
	defer rs.lk.Unlock()

	delete(rs.m, req)
	delete(rs.hm, req)
}

func encodeGob(obj interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeGob(encoded []byte) (map[string]interface{}, error) {
	buf := bytes.NewBuffer(encoded)
	dec := gob.NewDecoder(buf)
	var out map[string]interface{}
	err := dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// encode uses the given block cipher (in CTR mode) to encrypt the
// data, along with a hash, returning the iv and the ciphertext. What
// is returned looks like:
//
//   encrypted(salt + sessionData) + iv + hmac
//
func encode(block cipher.Block, hmac hash.Hash, data []byte) ([]byte, error) {

	buf := bytes.NewBuffer(nil)

	salt := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	buf.Write(salt)
	buf.Write(data)

	session := buf.Bytes()

	iv := make([]byte, block.BlockSize())
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(session, session)

	buf.Write(iv)
	hmac.Write(buf.Bytes())
	buf.Write(hmac.Sum(nil))

	return buf.Bytes(), nil
}

func encodeCookie(content interface{}, encKey, hmacKey []byte) (string, []byte, error) {
	encodedGob, err := encodeGob(content)
	if err != nil {
		return "", nil, err
	}

	gobHash := sha1.New()
	gobHash.Write(encodedGob)

	aesCipher, err := aes.NewCipher(encKey)
	if err != nil {
		return "", nil, err
	}

	hmacHash := hmac.New(sha256.New, hmacKey)

	sessionBytes, err := encode(aesCipher, hmacHash, encodedGob)
	if err != nil {
		return "", nil, err
	}

	return base64.StdEncoding.EncodeToString(sessionBytes), gobHash.Sum(nil), nil
}

// decode uses the given block cipher (in CTR mode) to decrypt the
// data, and validate the hash.  If hash validation fails, an error is
// returned.
func decode(block cipher.Block, hmac hash.Hash, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 2*block.BlockSize()+hmac.Size() {
		return nil, LenError
	}

	receivedHmac := ciphertext[len(ciphertext)-hmac.Size():]
	ciphertext = ciphertext[:len(ciphertext)-hmac.Size()]

	hmac.Write(ciphertext)
	if subtle.ConstantTimeCompare(hmac.Sum(nil), receivedHmac) != 1 {
		return nil, HashError
	}

	// split the iv and session bytes
	iv := ciphertext[len(ciphertext)-block.BlockSize():]
	session := ciphertext[:len(ciphertext)-block.BlockSize()]

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(session, session)

	// skip past the iv
	session = session[block.BlockSize():]

	return session, nil
}

func decodeCookie(encoded string, encKey, hmacKey []byte) (map[string]interface{}, []byte, error) {
	sessionBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, nil, err
	}
	aesCipher, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, nil, err
	}

	hmacHash := hmac.New(sha256.New, hmacKey)
	gobBytes, err := decode(aesCipher, hmacHash, sessionBytes)
	if err != nil {
		return nil, nil, err
	}

	gobHash := sha1.New()
	gobHash.Write(gobBytes)

	session, err := decodeGob(gobBytes)
	if err != nil {
		log.Printf("decodeGob: %s\n", err)
		return nil, nil, err
	}
	return session, gobHash.Sum(nil), nil
}

func (s sessionResponseWriter) WriteHeader(code int) {
	if atomic.AddInt32(&s.wroteHeader, 1) == 1 {
		origCookie, err := s.req.Cookie(s.h.CookieName)
		var origCookieVal string
		if err != nil {
			origCookieVal = ""
		} else {
			origCookieVal = origCookie.Value
		}

		session := s.h.RS.Get(s.req)
		if len(session) == 0 {
			// if we have an empty session, but the
			// request didn't start out that way, we
			// assume the user wants us to clear the
			// session
			if origCookieVal != "" {
				//log.Println("clearing cookie")
				var cookie http.Cookie
				cookie.Name = s.h.CookieName
				cookie.Value = ""
				cookie.Path = "/"
				// a cookie is expired by setting it
				// with an expiration time in the past
				cookie.Expires = time.Unix(0, 0).UTC()
				http.SetCookie(s, &cookie)
			}
			goto write
		}
		encoded, gobHash, err := encodeCookie(session, s.h.encKey, s.h.hmacKey)
		if err != nil {
			log.Printf("createCookie: %s\n", err)
			goto write
		}

		if bytes.Equal(gobHash, s.h.RS.getHash(s.req)) {
			//log.Println("not re-setting identical cookie")
			goto write
		}

		var cookie http.Cookie
		cookie.Name = s.h.CookieName
		cookie.Value = encoded
		cookie.Path = s.h.CookiePath
		cookie.HttpOnly = s.h.RS.HttpOnly
		cookie.Secure = s.h.RS.Secure
		http.SetCookie(s, &cookie)
	}
write:
	s.ResponseWriter.WriteHeader(code)
}

func (s sessionResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, _ := s.ResponseWriter.(http.Hijacker)
	return hijacker.Hijack()
}

func (h *SessionHandler) getCookieSession(req *http.Request) (map[string]interface{}, []byte) {
	cookie, err := req.Cookie(h.CookieName)
	if err != nil {
		//log.Printf("getCookieSesh: '%#v' not found\n",
		//	h.CookieName)
		return map[string]interface{}{}, nil
	}
	session, gobHash, err := decodeCookie(cookie.Value, h.encKey, h.hmacKey)
	if err != nil {
		log.Printf("decodeCookie: %s\n", err)
		return map[string]interface{}{}, nil
	}

	return session, gobHash
}

func (h *SessionHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// get our session a little early, so that we can add our
	// authentication information to it if we get some
	session, gobHash := h.getCookieSession(req)

	h.RS.Set(req, session, gobHash)
	defer h.RS.Clear(req)

	sessionWriter := sessionResponseWriter{rw, h, req, 0}
	h.Handler.ServeHTTP(sessionWriter, req)
}

func (this *Server) NewSessionHandler(key string, rs *RequestSessions) *SessionHandler {
	// sha1 sums are 20 bytes long.  we use the first 16 bytes as
	// the aes key.
	encHash := sha1.New()
	encHash.Write([]byte(key))
	encHash.Write([]byte("-encryption"))
	hmacHash := sha1.New()
	hmacHash.Write([]byte(key))
	hmacHash.Write([]byte("-hmac"))

	// if the user hasn't specified a session handler, use the
	// package's default one
	if rs == nil {
		rs = Session
	}

	return &SessionHandler{
		Handler:    this,
		CookieName: "session",
		CookiePath: "/",
		RS:         rs,
		encKey:     encHash.Sum(nil)[:blockSize],
		hmacKey:    hmacHash.Sum(nil)[:blockSize],
	}
}

/* Logging
   ------------------------------ */

func (this *Server) SetLogger(logger *log.Logger) {
	this.Logger = logger
}

func SetLogger(logger *log.Logger) {
	mainServer.Logger = logger
}

/* Templating |-- Using html/template library built into golang http://golang.org/pkg/html/template/ --|
   ------------------------------ */

func (this *Server) Template(w http.ResponseWriter) (templ Template, err error) {
	if w == nil {
		log.Printf("Template Error: %v", err.Error())
		return
	}
	templ.Writer = w
	templ.Bag = make(map[string]interface{})
	return
}

func (t Template) SinglePage(file_path string) (err error) {
	if t.Bag == nil {
		t.Bag = make(map[string]interface{})
	}
	if len(file_path) != 0 {
		t.Template = file_path
	}

	tmpl := template.Must(template.ParseFiles(t.Template))

	err = tmpl.Execute(t.Writer, t.Bag)

	return
}

func (t Template) DisplayTemplate() (err error) {
	if t.Layout == "" {
		t.Layout = "layout.html"
	}
	if t.Bag == nil {
		t.Bag = make(map[string]interface{})
	}

	templ := template.Must(template.ParseFiles(t.Layout, t.Template))
	err = templ.Execute(t.Writer, t.Bag)

	return
}

func (t Template) DisplayMultiple(templates []string) (err error) {
	if t.Layout == "" {
		t.Layout = "layout.html"
	}
	if t.Bag == nil {
		t.Bag = make(map[string]interface{})
	}

	templ := template.Must(template.ParseFiles(t.Layout))
	for _, filename := range templates {
		templ.ParseFiles(filename)
	}
	err = templ.Execute(t.Writer, t.Bag)

	return
}

/* Contextual structs for simpler request/response (AppEngine failure)
   ------------------------------------- */

type Context struct {
	Request *http.Request
	Params  map[string]string
	Server  *Server
	ResponseWriter
}

type ResponseWriter interface {
	Header() http.Header
	WriteHeader(status int)
	Write(data []byte) (n int, err error)
	Close()
}

func (ctx *Context) WriteString(content string) {
	ctx.ResponseWriter.Write([]byte(content))
}

func (ctx *Context) Abort(status int, body string) {
	ctx.ResponseWriter.WriteHeader(status)
	ctx.ResponseWriter.Write([]byte(body))
}

func (ctx *Context) Redirect(status int, url_ string) {
	ctx.ResponseWriter.Header().Set("Location", url_)
	ctx.ResponseWriter.WriteHeader(status)
	ctx.ResponseWriter.Write([]byte("Redirecting to: " + url_))
}

func (ctx *Context) NotModified() {
	ctx.ResponseWriter.WriteHeader(304)
}

func (ctx *Context) NotFound(message string) {
	ctx.ResponseWriter.WriteHeader(404)
	ctx.ResponseWriter.Write([]byte(message))
}

//Sets the content type by extension, as defined in the mime package. 
//For example, ctx.ContentType("json") sets the content-type to "application/json"
func (ctx *Context) ContentType(ext string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	ctype := mime.TypeByExtension(ext)
	if ctype != "" {
		ctx.Header().Set("Content-Type", ctype)
	}
}

func (ctx *Context) SetHeader(hdr string, val string, unique bool) {
	if unique {
		ctx.Header().Set(hdr, val)
	} else {
		ctx.Header().Add(hdr, val)
	}
}

//Sets a cookie -- duration is the amount of time in seconds. 0 = forever
func (ctx *Context) SetCookie(name string, value string, age int64) {
	var utctime time.Time
	if age == 0 {
		// 2^31 - 1 seconds (roughly 2038)
		utctime = time.Unix(2147483647, 0)
	} else {
		utctime = time.Unix(time.Now().Unix()+age, 0)
	}
	cookie := fmt.Sprintf("%s=%s; expires=%s", name, value, webTime(utctime))
	ctx.SetHeader("Set-Cookie", cookie, false)
}

func getCookieSig(key string, val []byte, timestamp string) string {
	hm := hmac.New(sha1.New, []byte(key))

	hm.Write(val)
	hm.Write([]byte(timestamp))

	hex := fmt.Sprintf("%02x", hm.Sum(nil))
	return hex
}

func (ctx *Context) SetSecureCookie(name string, val string, age int64) {
	//base64 encode the val
	if len(ctx.Server.Config.CookieSecret) == 0 {
		ctx.Server.Logger.Println("Secret Key for secure cookies has not been set. Please assign a cookie secret to web.Config.CookieSecret.")
		return
	}
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write([]byte(val))
	encoder.Close()
	vs := buf.String()
	vb := buf.Bytes()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	sig := getCookieSig(ctx.Server.Config.CookieSecret, vb, timestamp)
	cookie := strings.Join([]string{vs, timestamp, sig}, "|")
	ctx.SetCookie(name, cookie, age)
}
func (ctx *Context) GetSecureCookie(name string) (string, bool) {
	for _, cookie := range ctx.Request.Cookies() {
		if cookie.Name != name {
			continue
		}

		parts := strings.SplitN(cookie.Value, "|", 3)

		val := parts[0]
		timestamp := parts[1]
		sig := parts[2]

		if getCookieSig(ctx.Server.Config.CookieSecret, []byte(val), timestamp) != sig {
			return "", false
		}

		ts, _ := strconv.ParseInt(timestamp, 0, 64)

		if time.Now().Unix()-31*86400 > ts {
			return "", false
		}

		buf := bytes.NewBufferString(val)
		encoder := base64.NewDecoder(base64.StdEncoding, buf)

		res, _ := ioutil.ReadAll(encoder)
		return string(res), true
	}
	return "", false
}

// small optimization: cache the context type instead of repeteadly calling reflect.Typeof
var contextType reflect.Type

var exeFile string

// default
func defaultStaticDir() string {
	root, _ := path.Split(exeFile)
	return path.Join(root, "static")
}

func init() {
	contextType = reflect.TypeOf(Context{})
	//find the location of the exe file
	arg0 := path.Clean(os.Args[0])
	wd, _ := os.Getwd()
	if strings.HasPrefix(arg0, "/") {
		exeFile = arg0
	} else {
		//TODO for robustness, search each directory in $PATH
		exeFile = path.Join(wd, arg0)
	}
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

func webTime(t time.Time) string {
	ftime := t.Format(time.RFC1123)
	if strings.HasSuffix(ftime, "UTC") {
		ftime = ftime[0:len(ftime)-3] + "GMT"
	}
	return ftime
}

func dirExists(dir string) bool {
	d, e := os.Stat(dir)
	switch {
	case e != nil:
		return false
	case !d.IsDir():
		return false
	}

	return true
}

func fileExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func Urlencode(data map[string]string) string {
	var buf bytes.Buffer
	for k, v := range data {
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(url.QueryEscape(v))
		buf.WriteByte('&')
	}
	s := buf.String()
	return s[0 : len(s)-1]
}

func Serve404(w http.ResponseWriter, error string) {
	tmpl, _ := mainServer.Template(w)
	tmpl.Bag["Error"] = error
	_ = tmpl.SinglePage("templates/404.html")
}
