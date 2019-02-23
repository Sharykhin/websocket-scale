package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":3000", "web server addr")
	flag.Parse()

	tmpl := template.Must(template.ParseFiles(  "templates/index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.Execute(w, nil)
		if err != nil {
			log.Fatalf("could not execute template: %v", err)
		}
	})
	fmt.Printf("Server is running on %s \n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
