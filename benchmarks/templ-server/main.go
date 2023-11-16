package main

import (
	"log"
	"net/http"

	"github.com/a-h/templ"
)

type Person struct {
	Name  string
	Email string
}

func main() {
	t := Render(Person{
		Name:  "Luiz Bonfa",
		Email: "luiz@example.com",
	})
	err := http.ListenAndServe("localhost:8080", templ.Handler(t))
	if err != nil {
		log.Fatal(err)
	}
}
