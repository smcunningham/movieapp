package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"

	"movieapp/db"

	"github.com/gomodule/redigo/redis"
)

const (
	templateFileDir = "./templates/"
	spaceFormat     = "%20"
)

var pgDb *sql.DB
var loginTmpl, homeTmpl, profileTmpl, searchTmpl, shareTmpl *template.Template
var cache redis.Conn

func init() {
	// initialize db
	pgDb = db.OpenDB()

	// initialize redis connection for sessions and assign to cache var
	conn, err := redis.DialURL("redis://localhost")
	if err != nil {
		panic(err)
	}
	cache = conn

	// create and cache templates
	homeTmpl, loginTmpl, profileTmpl, searchTmpl, shareTmpl =
		template.Must(template.ParseFiles(templateFileDir+"home.html")),
		template.Must(template.ParseFiles(templateFileDir+"login.html")),
		template.Must(template.ParseFiles(templateFileDir+"profile.html")),
		template.Must(template.ParseFiles(templateFileDir+"search.html")),
		template.Must(template.ParseFiles(templateFileDir+"share.html"))
	fmt.Printf("Created templates: %v\n%v\n%v\n%v\n%v\n", homeTmpl, loginTmpl, profileTmpl, searchTmpl, shareTmpl)
}

func main() {
	defer pgDb.Close()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Handle templates
	http.HandleFunc("/", loginHandler)
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/searchActive", searchActiveHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/share", shareHandler)
	http.HandleFunc("/signup", signupHandler)

	log.Println("listening...")
	http.ListenAndServe(":3000", nil)
}

func consoleLog(name string, r *http.Request) {
	fmt.Printf("%s - Request Method: %s Request URL: %s\n", name, r.Method, r.URL.EscapedPath())
}
