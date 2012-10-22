package controllers

import (
	//	"fmt"
	"fmt"
	"gophers/plate"
	"net/http"
	"time"
)

type Person struct {
	First     string
	Last      string
	Email     string
	DateAdded time.Time
}

func Index(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintf(w, "Hello world!")
	plate.Serve404(w, "")
}

func World(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello world!")
}

func ByURL(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	name := params.Get(":name")
	fmt.Fprintf(w, "Hello %v!", name)
}

func PostWorld(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	fmt.Fprintf(w, "Hello %v!", name)
}

func DeleteVegetables(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Let's delete the veggies!")
}

func PutFruit(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Let's add some fruit!")
}

func PatchMeat(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Let's patch the meat!")
}

func Sensi(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "This route is a little sensitive")
}

func Secure(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "You must be authenticated or something :|")
}

func PeopleJson(w http.ResponseWriter, r *http.Request) {
	ppl := make([]Person, 0)

	p1 := Person{
		First:     "Alex",
		Last:      "Ninneman",
		Email:     "alex@ninneman.org",
		DateAdded: time.Now(),
	}
	ppl = append(ppl, p1)

	p2 := Person{
		First:     "Jessica",
		Last:      "Janiuk",
		Email:     "jessica.janiuk@gmail.com",
		DateAdded: time.Now(),
	}
	ppl = append(ppl, p2)

	plate.ServeJson(w, ppl)
}

func PeopleXml(w http.ResponseWriter, r *http.Request) {
	ppl := make([]Person, 0)

	p1 := Person{
		First:     "Alex",
		Last:      "Ninneman",
		Email:     "alex@ninneman.org",
		DateAdded: time.Now(),
	}
	ppl = append(ppl, p1)

	p2 := Person{
		First:     "Jessica",
		Last:      "Janiuk",
		Email:     "jessica.janiuk@gmail.com",
		DateAdded: time.Now(),
	}
	ppl = append(ppl, p2)

	plate.ServeXml(w, ppl)
}

func PeopleFormatted(w http.ResponseWriter, r *http.Request) {
	ppl := make([]Person, 0)

	p1 := Person{
		First:     "Alex",
		Last:      "Ninneman",
		Email:     "alex@ninneman.org",
		DateAdded: time.Now(),
	}
	ppl = append(ppl, p1)

	p2 := Person{
		First:     "Jessica",
		Last:      "Janiuk",
		Email:     "jessica.janiuk@gmail.com",
		DateAdded: time.Now(),
	}
	ppl = append(ppl, p2)

	plate.ServeFormatted(w, r, ppl)
}
