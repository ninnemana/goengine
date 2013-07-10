package plate

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
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
		Writer:  w,
		Bag:     make(map[string]interface{}),
		FuncMap: template.FuncMap{},
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

	tmpl, err := template.ParseFiles(t.Template)
	if err != nil {
		return err
	}
	tmpl.Funcs(t.FuncMap)
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

	templ, err := template.ParseFiles(t.Layout, t.Template)
	if err != nil {
		return err
	}
	templ.Funcs(t.FuncMap)

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
	t.Layout = t.Layout

	templ, err := template.ParseFiles(dir + "/" + t.Layout)
	if err != nil {
		return err
	}
	templ.Funcs(t.FuncMap)
	for _, filename := range templates {
		_, err = templ.ParseFiles(dir + "/" + filename)
	}

	err = templ.Execute(t.Writer, t.Bag)
	if err != nil {
		log.Println(err)
	}

	return
}

func SetTemplate(t *Template) {
	tmpl = t
}

func (t *Template) ParseFile(file string, override bool) error {
	dir := ""
	var err error
	if !override {
		dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	var tmpl *template.Template
	if t.HtmlTemplate == nil {

		tmpl, err = template.New(file).Funcs(t.FuncMap).ParseFiles(dir + "/" + file)
		if err != nil {
			return err
		}
		tmpl.Funcs(t.FuncMap)
	} else {
		// Make sure the FuncMap is added to the template before parsing the new file
		t.HtmlTemplate.Funcs(t.FuncMap)
		tmpl, err = t.HtmlTemplate.ParseFiles(dir + "/" + file)
		if err != nil {
			return err
		}
	}
	t.HtmlTemplate = tmpl

	return err
}

func (t Template) Display(w http.ResponseWriter) error {

	t.Writer = w
	if t.HtmlTemplate == nil {
		if t.Template == "" {
			return errors.New("No template files defined")
		}

		dir, err := os.Getwd()
		if err != nil {
			return err
		}

		tmpl, err := template.ParseFiles(dir + "/" + t.Template)

		if err != nil {
			log.Println(err)
			return err
		}
		tmpl.Funcs(t.FuncMap)

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

func NewTemplate(w http.ResponseWriter) *Template {
	server := NewServer()

	tmpl, err := GetTemplate()
	if err != nil {
		tmpl, err = server.Template(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return tmpl
		}
	}
	return tmpl
}
