package helpers

import(
	"net/http"
	"html/template"
)

func DisplayTemplate(name string, filename string, w http.ResponseWriter, x map[string]interface{}){

	t := template.Must(template.ParseFiles("layout.html",filename))
	
	if err := t.Execute(w, x); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}