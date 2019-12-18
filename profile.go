package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func profileHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("profileHandler", r)

	if err := profileTmpl.ExecuteTemplate(w, "profile", td); err != nil {
		log.Printf("Failed to execute template: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func shareHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("shareHandler", r)

	if err := shareTmpl.ExecuteTemplate(w, "share", td); err != nil {
		log.Printf("Failed to execute template: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Post request handler for adding user to database
func signupHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("signupHandler", r)

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
		0, // serial column in db so can be any number
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
