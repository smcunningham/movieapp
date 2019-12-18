package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// Contains data about the user, passed to HTML templates
var td TemplateData

// Creds holds new user infornation
type Creds struct {
	Password  string `json:"password", db:"pword"`
	Username  string `json:"username", db:"uname"`
	Firstname string `json:"firstname", db:"fname"`
	Lastname  string `json:"lastname", db:"lname"`
	Email     string `json:"email", db:"email"`
}

// Login page
func loginHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("loginHandler", r)

	if err := loginTmpl.ExecuteTemplate(w, "login", nil); err != nil {
		log.Printf("Failed to execute template: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Home page after login
func homeHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("homeHandler", r)

	creds := Creds{
		Username: r.FormValue("username"),
		Password: r.FormValue("pass"),
	}
	fmt.Printf("Username: %s\nPassword: %s\n", creds.Username, creds.Password)

	if login(creds) {
		fmt.Println("Login Successful")
		td.User = creds.Username

		// Execute login template
		if err := homeTmpl.ExecuteTemplate(w, "home", td); err != nil {
			log.Printf("Failed to execute template: %+v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// Login failed, return to login page
	fmt.Println("Login Failed")
	loginHandler(w, r)
}

// Login function
func login(creds Creds) bool {
	// Try to find username in user table
	loginStmt := `SELECT pword FROM users WHERE uname=$1`
	row := pgDb.QueryRow(loginStmt, creds.Username)

	storedCreds := Creds{}
	switch err := row.Scan(&storedCreds.Password); err {
	case sql.ErrNoRows:
		fmt.Println("No rows returned, username not found!")
		return false
	case nil:
		// Compare passwords
		if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
			// No match
			fmt.Println("Incorrect Password!")
			return false
		}
		return true
	default:
		// All other errors
		fmt.Printf("Error during login: %s\n", err)
		return false
	}
}
