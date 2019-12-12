package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"

	"movieapp/db"

	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/bcrypt"
)

const (
	staticFileDir   = "./static/"
	templateFileDir = "./templates/"

	apiURL        = "https://movie-database-imdb-alternative.p.rapidapi.com/?r=json&s="
	imdbSearchURL = "https://www.imdb.com/title/"
	apiHostHeader = "x-rapidapi-host"
	apiHostValue  = "movie-database-imdb-alternative.p.rapidapi.com"
	apiKeyHeader  = "x-rapidapi-key"
	apiKeyValue   = "8d2e6b2afbmsh6fb117b4fc923a7p1d12b2jsnf5603af8be8b"
	spaceFormat   = "%20"
)

// Contains data about the user, passed to HTML templates
var td TemplateData

var un, pass string
var pgDb *sql.DB
var loginTmpl, homeTmpl, profileTmpl, searchTmpl, shareTmpl *template.Template

var cache redis.Conn

// TemplateData is passed to the template
type TemplateData struct {
	MovieData    []Data
	TotalResults string
	User         string
}

// Search holds generic search data returned from the movie API
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

// Creds holds new user infornation
type Creds struct {
	Password  string `json:"password", db:"pword"`
	Username  string `json:"username", db:"uname"`
	Firstname string `json:"firstname", db:"fname"`
	Lastname  string `json:"lastname", db:"lname"`
	Email     string `json:"email", db:"email"`
}

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
	homeTmpl, loginTmpl, profileTmpl, searchTmpl, shareTmpl = template.Must(template.ParseFiles(templateFileDir+"home.html")),
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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	printData("loginHandler", r)
	if err := loginTmpl.ExecuteTemplate(w, "login", nil); err != nil {
		log.Printf("Failed to execute template: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	printData("homeHandler", r)

	creds := Creds{
		Username: r.FormValue("username"),
		Password: r.FormValue("pass"),
	}
	fmt.Printf("Username: %s\nPassword: %s\n", creds.Username, creds.Password)

	if login(creds) {
		fmt.Println("Login Successful")
		td.User = creds.Username

		if err := homeTmpl.ExecuteTemplate(w, "home", td); err != nil {
			log.Printf("Failed to execute template: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	fmt.Println("Login Failed")
	loginHandler(w, r)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	printData("profileHandler", r)
	if err := profileTmpl.ExecuteTemplate(w, "profile", td); err != nil {
		log.Printf("Failed to execute template: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func searchActiveHandler(w http.ResponseWriter, r *http.Request) {
	printData("searchActiveHandler", r)

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
			w.WriteHeader(http.StatusBadRequest)
		}
		defer res.Body.Close()

		response, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(response))

		err = json.Unmarshal(response, &searchResults)
		if err != nil {
			fmt.Printf("%T\n%s\n%v\n", err, err, err)
			w.WriteHeader(http.StatusBadRequest)
		}

		td.MovieData = searchResults.Search
		td.TotalResults = searchResults.TotalResults

		if err := searchTmpl.ExecuteTemplate(w, "search", td); err != nil {
			log.Printf("Failed to execute template: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if err := searchTmpl.ExecuteTemplate(w, "search", td); err != nil {
		log.Printf("Failed to execute template: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	printData("searchHandler", r)

	if err := searchTmpl.ExecuteTemplate(w, "search", td); err != nil {
		log.Printf("Failed to execute template: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func shareHandler(w http.ResponseWriter, r *http.Request) {
	printData("shareHandler", r)
	if err := shareTmpl.ExecuteTemplate(w, "share", td); err != nil {
		log.Printf("Failed to execute template: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	printData("signupHandler", r)
	// Parse and decode request into 'creds' struct
	creds := Creds{}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		fmt.Printf("error decoding: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)

	if _, err := pgDb.Query(`INSERT INTO users values ($1, $2, $3, $4, $5, $6)`,
		0,
		creds.Username,
		string(hashedPassword),
		creds.Firstname,
		creds.Lastname,
		creds.Email); err != nil {
		fmt.Printf("error inserting new user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func printData(name string, r *http.Request) {
	fmt.Printf("%s - Request Method: %s Request URL: %s\n", name, r.Method, r.URL.EscapedPath())
}

func login(creds Creds) bool {
	loginStmt := `SELECT pword FROM users WHERE uname=$1`
	row := pgDb.QueryRow(loginStmt, creds.Username)

	storedCreds := Creds{}
	switch err := row.Scan(&storedCreds.Password); err {
	case sql.ErrNoRows:
		fmt.Println("No rows returned, username not found!")
		return false
	case nil:
		if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
			fmt.Println("Incorrect Password!")
			return false
		}
		return true
	default:
		fmt.Printf("Error during login: %s\n", err)
		return false
	}
}
