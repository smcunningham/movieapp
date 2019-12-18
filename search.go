package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	apiURL        = "https://movie-database-imdb-alternative.p.rapidapi.com/?r=json&s="
	imdbSearchURL = "https://www.imdb.com/title/"
	apiHostHeader = "x-rapidapi-host"
	apiHostValue  = "movie-database-imdb-alternative.p.rapidapi.com"
	apiKeyHeader  = "x-rapidapi-key"
	apiKeyValue   = "8d2e6b2afbmsh6fb117b4fc923a7p1d12b2jsnf5603af8be8b"
)

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

func searchActiveHandler(w http.ResponseWriter, r *http.Request) {
	consoleLog("searchActiveHandler", r)

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
	consoleLog("searchHandler", r)

	if err := searchTmpl.ExecuteTemplate(w, "search", td); err != nil {
		log.Printf("Failed to execute template: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
