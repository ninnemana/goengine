package helpers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type Template struct {
	Layout   string
	Template string
	Bag      map[string]interface{}
	Writer   http.ResponseWriter
}

type TemplateError struct {
	Message string
	Time    time.Time
}

func (e TemplateError) Error() string {
	return fmt.Sprintf("%v: %v", e.Time, e.Message)
}

func New(w http.ResponseWriter) (templ Template, err TemplateError) {
	if w == nil {
		err.Message = "Must pass ResponseWriter to New function"
	}
	templ.Writer = w
	templ.Bag = make(map[string]interface{})
	return
}

func (t Template) DisplayTemplate() {
	if t.Layout == "" {
		t.Layout = "layout.html"
	}
	if t.Bag == nil {
		t.Bag = make(map[string]interface{})
	}
	templ := template.Must(template.ParseFiles(t.Layout, t.Template))

	if err := templ.Execute(t.Writer, t.Bag); err != nil {
		http.Error(t.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func (t Template) DisplayMultiple(templates []string) {
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
	if err := templ.Execute(t.Writer, t.Bag); err != nil {
		http.Error(t.Writer, err.Error(), http.StatusInternalServerError)
	}

}
