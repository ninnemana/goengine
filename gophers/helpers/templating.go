package helpers

import (
	//"net/http"
	//"html/template"
	"fmt"
)

func DisplayTemplate() {
	fmt.Println("I am annoyed as hell")

}

/*func DisplayTemplate(ctx *WebContext) {
	if ctx.layout == "" {
		ctx.layout = "layout.html"
	}
	t := template.Must(template.ParseFiles(ctx.layout, "templates/"+ctx.template))

	if err := t.Execute(w, x); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}*/
