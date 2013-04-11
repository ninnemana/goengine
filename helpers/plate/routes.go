package plate

import (
	"net/http"
	"regexp"
	"strings"
)

type Route struct {
	method      string
	regex       *regexp.Regexp
	params      map[int]string
	handler     http.HandlerFunc
	sensitive   bool
	filters     []http.HandlerFunc
	contenttype string
	unfiltered  bool // this will ignore all global filters on this route
}

func (this *Route) Sensitive() *Route {
	this.sensitive = true
	return this
}

// Allows overriding the default content type header
func (this *Route) ContentType(contenttype string) *Route {
	this.contenttype = contenttype
	return this
}

func (this *Route) NoFilter() *Route {
	this.unfiltered = true
	return this
}

// Add middleware filter to specific route
func (this *Route) AddFilter(filter http.HandlerFunc) {
	this.filters = append(this.filters, filter)
}

func (this *Route) FilterParam(param string, filter http.HandlerFunc) {
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
