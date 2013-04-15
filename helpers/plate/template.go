package plate

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	tmpl *Template
)

type Template struct {
	Layout       string
	Template     string
	Bag          map[string]interface{}
	Writer       http.ResponseWriter
	FuncMap      template.FuncMap
	HtmlTemplate *template.Template
}

/* Templating |-- Using html/template library built into golang http://golang.org/pkg/html/template/ --|
   ------------------------------ */

func (this *Server) Template(w http.ResponseWriter) (templ *Template, err error) {
	if w == nil {
		log.Printf("Template Error: %v", err.Error())
		return
	}
	templ = &Template{
		Writer: w,
		Bag:    make(map[string]interface{}),
	}

	return
}

func (t Template) SinglePage(file_path string) (err error) {

	dir, err := os.Getwd()
	if err != nil {
		return
	}

	if t.Bag == nil {
		t.Bag = make(map[string]interface{})
	}
	if len(file_path) != 0 {
		t.Template = dir + "/" + file_path
	}

	// the template name must match the first file it parses, but doesn't accept slashes
	// the following block ensures a match
	templateName := t.Template
	if strings.Index(templateName, "/") > -1 {
		tparts := strings.Split(templateName, "/")
		templateName = tparts[len(tparts)-1]
	}

	tmpl, err := template.New(templateName).Funcs(t.FuncMap).ParseFiles(t.Template)
	if err != nil {
		return err
	}
	err = tmpl.Execute(t.Writer, t.Bag)

	return
}

func (t Template) DisplayTemplate() (err error) {

	dir, err := os.Getwd()
	if err != nil {
		return
	}

	if t.Layout == "" {
		t.Layout = "layout.html"
	}
	// ensure proper pathing for layout layout files
	t.Layout = dir + "/" + t.Layout
	t.Template = dir + "/" + t.Template

	if t.Bag == nil {
		t.Bag = make(map[string]interface{})
	}

	// the template name must match the first file it parses, but doesn't accept slashes
	// the following block ensures a match
	templateName := t.Layout
	if strings.Index(templateName, "/") > -1 {
		tparts := strings.Split(templateName, "/")
		templateName = tparts[len(tparts)-1]
	}

	templ, err := template.New(templateName).Funcs(t.FuncMap).ParseFiles(t.Layout, t.Template)
	if err != nil {
		return err
	}

	err = templ.Execute(t.Writer, t.Bag)

	return
}

func (t Template) DisplayMultiple(templates []string) (err error) {

	dir, err := os.Getwd()
	if err != nil {
		return
	}

	if t.Layout == "" {
		t.Layout = "layout.html"
	}
	// ensure proper pathing for layout layout files
	t.Layout = dir + "/" + t.Layout

	if t.Bag == nil {
		t.Bag = make(map[string]interface{})
	}

	// the template name must match the first file it parses, but doesn't accept slashes
	// the following block ensures a match
	templateName := t.Layout
	if strings.Index(templateName, "/") > -1 {
		tparts := strings.Split(templateName, "/")
		templateName = tparts[len(tparts)-1]
	}

	templ, err := template.New(templateName).Funcs(t.FuncMap).ParseFiles(t.Layout)
	if err != nil {
		return err
	}
	for _, filename := range templates {
		templ.ParseFiles(dir + "/" + filename)
	}
	err = templ.Execute(t.Writer, t.Bag)

	return
}
func SetTemplate(t *Template) {
	tmpl = t
}

func (t *Template) ParseFile(file string) error {

	if t.HtmlTemplate == nil {
		var tmplName string
		if strings.Index(file, "/") > -1 {
			tparts := strings.Split(file, "/")
			tmplName = tparts[len(tparts)-1]
		}
		tmpl, err := template.New(tmplName).Funcs(t.FuncMap).ParseFiles(file)
		if err != nil {
			return err
		}
		t.HtmlTemplate = tmpl
		return nil
	}

	_, err := t.HtmlTemplate.ParseFiles(file)

	return err
}

func (t Template) Display(w http.ResponseWriter) error {

	t.Writer = w
	if t.HtmlTemplate == nil {
		if t.Template == "" {
			return errors.New("No template files defined")
		}

		tmplName := t.Template
		if strings.Index(t.Template, "/") > -1 {
			tparts := strings.Split(t.Template, "/")
			tmplName = tparts[len(tparts)-1]
		}
		tmpl, err := template.New(tmplName).Funcs(t.FuncMap).ParseFiles(t.Template)
		if err != nil {
			return err
		}
		t.HtmlTemplate = tmpl
		return nil
	}

	err := tmpl.HtmlTemplate.Execute(t.Writer, t.Bag)
	return err

}

func GetTemplate() (*Template, error) {
	if tmpl != nil {
		return tmpl, nil
	}

	return nil, errors.New("No template defined")
}
