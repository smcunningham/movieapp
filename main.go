package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"

	"movieapp/db"
	help "movieapp/helpers"
)

const (
	apiURL        = "https://movie-database-imdb-alternative.p.rapidapi.com/?r=json&s="
	apiHostHeader = "x-rapidapi-host"
	apiHostValue  = "movie-database-imdb-alternative.p.rapidapi.com"
	apiKeyHeader  = "x-rapidapi-key"
	apiKeyValue   = "8d2e6b2afbmsh6fb117b4fc923a7p1d12b2jsnf5603af8be8b"
	spaceFormat   = "%20"
)

var loginTmpl, homeTmpl, profileTmpl, searchTmpl, shareTmpl *template.Template
var staticFileDir = "./static/"
var templateFileDir = "./templates/"

var un, pass string

// TemplateData is passed to the template
type TemplateData struct {
	MovieData    []Data
	TotalResults string
	User         string
}

type Search struct {
	Search       []Data `json:"Search"`
	TotalResults string `json:"totalResults"`
	Response     string `json:"Response"`
}

// Data holds the information about each movie returned in the search
type Data struct {
	ID     string `json:"imdbID"`
	Title  string `json:"Title"`
	Year   string `json:"Year"`
	Type   string `json:"Type"`
	ImgURL string `json:"Poster"`
}

var td TemplateData

func init() {
	// create and cache templates
	homeTmpl, loginTmpl, profileTmpl, searchTmpl, shareTmpl = template.Must(template.ParseFiles(templateFileDir+"home.html")),
		template.Must(template.ParseFiles(templateFileDir+"login.html")),
		template.Must(template.ParseFiles(templateFileDir+"profile.html")),
		template.Must(template.ParseFiles(templateFileDir+"search.html")),
		template.Must(template.ParseFiles(templateFileDir+"share.html"))
	fmt.Printf("Created templates: %v\n%v\n%v\n%v\n%v\n", homeTmpl, loginTmpl, profileTmpl, searchTmpl, shareTmpl)
}

func main() {
	db.Login()

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", loginHandler)
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/searchActive", searchActiveHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/share", shareHandler)

	log.Println("listening...")
	http.ListenAndServe(":3000", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	n := "loginHandler"
	if err := loginTmpl.ExecuteTemplate(w, "login", nil); err != nil {
		log.Printf("Failed to execute template: %+v", err)
	}
	printData(n, r)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	n := "homeHandler"
	printData(n, r)

	un = r.FormValue("username")
	pass = r.FormValue("pass")

	fmt.Printf("Username: %s\nPassword: %s\n", un, pass)

	unCheck := help.IsEmpty(un)
	passCheck := help.IsEmpty(pass)

	if unCheck || passCheck {
		fmt.Fprintf(w, "ErrorCode is -10 : Empty data found in the form.")
		return
	}

	td.User = un

	dbPwd := "1234"
	dbUn := "stevanc"

	if dbPwd == pass && dbUn == un {
		fmt.Println("Login Successful")
		if err := homeTmpl.ExecuteTemplate(w, "home", td); err != nil {
			log.Printf("Failed to execute template: %+v", err)
		}
	} else {
		fmt.Println("Login Failed")
		fmt.Fprintln(w, "Login Failed!")
	}
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	n := "profileHandler"
	if err := profileTmpl.ExecuteTemplate(w, "profile", td); err != nil {
		log.Printf("Failed to execute template: %+v", err)
	}
	printData(n, r)
}

func searchActiveHandler(w http.ResponseWriter, r *http.Request) {
	n := "searchHandler"
	printData(n, r)

	var searchArr []string
	var searchResults Search

	s := r.FormValue("searchword")
	search := strings.Trim(s, " ")

	if strings.Contains(search, " ") {
		searchArr = strings.Split(search, " ")
	} else {
		searchArr = append(searchArr, search)
	}

	if len(searchArr) > 0 {
		// if not empty, search
		search = strings.Join(searchArr, " ")
		fmt.Printf("Search string: %s\n", search)

		search = strings.Replace(search, " ", spaceFormat, -1)
		fmt.Printf("API search string : %s\n", search)

		url := apiURL + search

		req, _ := http.NewRequest("GET", url, nil)

		req.Header.Add(apiHostHeader, apiHostValue)
		req.Header.Add(apiKeyHeader, apiKeyValue)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("%T\n%s\n%v\n", err, err, err)
		}
		defer res.Body.Close()

		response, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(response))

		err = json.Unmarshal(response, &searchResults)
		if err != nil {
			fmt.Printf("%T\n%s\n%v\n", err, err, err)
		}

		td.MovieData = searchResults.Search
		td.TotalResults = searchResults.TotalResults

		if err := searchTmpl.ExecuteTemplate(w, "search", td); err != nil {
			log.Printf("Failed to execute template: %+v", err)
		}
		return
	}

	if err := searchTmpl.ExecuteTemplate(w, "search", td); err != nil {
		log.Printf("Failed to execute template: %+v", err)
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	n := "searchHandler"
	printData(n, r)

	if err := searchTmpl.ExecuteTemplate(w, "search", td); err != nil {
		log.Printf("Failed to execute template: %+v", err)
	}
}

func shareHandler(w http.ResponseWriter, r *http.Request) {
	n := "shareHandler"
	if err := shareTmpl.ExecuteTemplate(w, "share", td); err != nil {
		log.Printf("Failed to execute template: %+v", err)
	}
	printData(n, r)
}

func printData(name string, r *http.Request) {
	fmt.Printf("%s - Request Method: %s Request URL: %s\n", name, r.Method, r.URL.EscapedPath())
}
