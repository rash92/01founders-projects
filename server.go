package main

import (
	"crypto/tls"
	"forum/controller"
	"forum/dbmanagement"
	"html/template"
	"log"
	"net/http"
)

var tmpl *template.Template

func init() {
	tmpl = template.Must(template.ParseGlob("static/*.html"))
}

func main() {
	path := "static"
	fs := http.FileServer(http.Dir(path))

	mux := http.NewServeMux()

	cert, _ := tls.LoadX509KeyPair("localhost.crt", "localhost.key")

	s := &http.Server{
		Addr:    ":9000",
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	}

	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		controller.Home(w, r, tmpl)
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		controller.Login(w, r, tmpl)
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		controller.Register(w, r, tmpl)
	})

	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		controller.Register(w, r, tmpl)
	})

	dbmanagement.CreateDatabaseWithTables()
	dbmanagement.DisplayAllUsers()
	log.Fatal(s.ListenAndServeTLS("", ""))
}
